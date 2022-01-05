// Copyright 2016 The Cockroach Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package almost_crdb

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/lib/pq"
)

// SlightlyAdaptedExecuteTx is copied from crdb/tx.go (and crdb.error.go) with a twist: the tx function here accepts a tx
// and can return results. These results are then propagated by SlightlyAdaptedExecuteTx
func SlightlyAdaptedExecuteTx(ctx context.Context, db *sql.DB, opts *sql.TxOptions, fn func(tx SlightlyAdaptedTx) (interface{}, error)) (interface{}, error) {
	rawTx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	tx := slightlyAdapterStdlibTxnAdapter{rawTx}
	defer func() {
		if err == nil {
			// Ignore commit errors. The tx has already been committed by RELEASE.
			_ = tx.Commit(ctx)
		} else {
			// We always need to execute a Rollback() so sql.DB releases the
			// connection.
			_ = tx.Rollback(ctx)
		}
	}()
	// Specify that we intend to retry this txn in case of CockroachDB retryable
	// errors.
	if _, err = tx.SlightlyAdaptedExec(ctx, "SAVEPOINT cockroach_restart"); err != nil {
		return nil, err
	}

	for {
		released := false
		result, err := fn(tx)
		if err == nil {
			// RELEASE acts like COMMIT in CockroachDB. We use it since it gives us an
			// opportunity to react to retryable errors, whereas tx.Commit() doesn't.
			released = true
			if _, err = tx.SlightlyAdaptedExec(ctx, "RELEASE SAVEPOINT cockroach_restart"); err == nil {
				return result, nil
			}
		}
		// We got an error; let's see if it's a retryable one and, if so, restart.
		if !errIsRetryable(err) {
			if released {
				err = newAmbiguousCommitError(err)
			}
			return nil, err
		}

		if _, retryErr := tx.SlightlyAdaptedExec(ctx, "ROLLBACK TO SAVEPOINT cockroach_restart"); retryErr != nil {
			return nil, newTxnRestartError(retryErr, err)
		}
	}
}

type SlightlyAdaptedTx interface {
	// SlightlyAdaptedExec has been changed (from crdb.Tx#Exec) to return results
	SlightlyAdaptedExec(context.Context, string, ...interface{}) (sql.Result, error)
	Commit(context.Context) error
	Rollback(context.Context) error
}

type slightlyAdapterStdlibTxnAdapter struct {
	tx *sql.Tx
}

var _ SlightlyAdaptedTx = slightlyAdapterStdlibTxnAdapter{}

func (tx slightlyAdapterStdlibTxnAdapter) SlightlyAdaptedExec(ctx context.Context, q string, args ...interface{}) (sql.Result, error) {
	return tx.tx.ExecContext(ctx, q, args...)
}

// Commit is part of the tx interface.
func (tx slightlyAdapterStdlibTxnAdapter) Commit(context.Context) error {
	return tx.tx.Commit()
}

// Commit is part of the tx interface.
func (tx slightlyAdapterStdlibTxnAdapter) Rollback(context.Context) error {
	return tx.tx.Rollback()
}

func errIsRetryable(err error) bool {
	// We look for either:
	//  - the standard PG errcode SerializationFailureError:40001 or
	//  - the Cockroach extension errcode RetriableError:CR000. This extension
	//    has been removed server-side, but support for it has been left here for
	//    now to maintain backwards compatibility.
	code := errCode(err)
	return code == "CR000" || code == "40001"
}

func errCode(err error) string {
	switch t := errorCause(err).(type) {
	case *pq.Error:
		return string(t.Code)

	case errWithSQLState:
		return t.SQLState()

	default:
		return ""
	}
}

// errorCause returns the original cause of the error, if possible. An
// error has a proximate cause if it's type is compatible with Go's
// errors.Unwrap() or pkg/errors' Cause(); the original cause is the
// end of the causal chain.
func errorCause(err error) error {
	for err != nil {
		if c, ok := err.(interface{ Cause() error }); ok {
			err = c.Cause()
		} else if c, ok := err.(interface{ Unwrap() error }); ok {
			err = c.Unwrap()
		} else {
			break
		}
	}
	return err
}

type txError struct {
	cause error
}

type errWithSQLState interface {
	SQLState() string
}

// Error implements the error interface.
func (e *txError) Error() string { return e.cause.Error() }

// Cause implements the pkg/errors causer interface.
func (e *txError) Cause() error { return e.cause }

// Unwrap implements the go error causer interface.
func (e *txError) Unwrap() error { return e.cause }

// AmbiguousCommitError represents an error that left a transaction in an
// ambiguous state: unclear if it committed or not.
type AmbiguousCommitError struct {
	txError
}

func newAmbiguousCommitError(err error) *AmbiguousCommitError {
	return &AmbiguousCommitError{txError{cause: err}}
}

// TxnRestartError represents an error when restarting a transaction. `cause` is
// the error from restarting the txn and `retryCause` is the original error which
// triggered the restart.
type TxnRestartError struct {
	txError
	retryCause error
	msg        string
}

func newTxnRestartError(err error, retryErr error) *TxnRestartError {
	const msgPattern = "restarting txn failed. ROLLBACK TO SAVEPOINT " +
		"encountered error: %s. Original error: %s."
	return &TxnRestartError{
		txError:    txError{cause: err},
		retryCause: retryErr,
		msg:        fmt.Sprintf(msgPattern, err, retryErr),
	}
}
