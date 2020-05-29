import datetime as dt
import json


def read_actions_over_time(filename: str):
    with open(filename) as f:
        data = json.load(f)
    if "Results" in data and "GroupedActionsOverTime" in data["Results"]:
        data = data["Results"]["GroupedActionsOverTime"]
    actions_over_time = []
    for key, value in data["Actions"].items():
        parsed_time = dt.datetime.fromisoformat(key.rstrip("Z"))
        actions_over_time.append((parsed_time, value))
    return sorted(actions_over_time, key=lambda a: a[0])
