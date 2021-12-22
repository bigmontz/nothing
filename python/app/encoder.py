from flask.json import JSONEncoder
from datetime import datetime


class AppJSONEncoder(JSONEncoder):
    def default(self, obj):
        try:
            if isinstance(obj, datetime) or isinstance():
                return obj.isoformat()
            iterable = iter(obj)
        except TypeError:
            pass
        else:
            return list(iterable)
        return JSONEncoder.default(self, obj)
