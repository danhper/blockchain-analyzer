from collections import defaultdict
import datetime as dt
import json
from typing import List, Tuple, Dict


def count_actions_over_time(actions: List[Tuple[dt.datetime, dict]]) -> Dict[str, int]:
    result = defaultdict(int)
    for _, actions_count in actions:
        for action, count in actions_count.items():
            result[action] += count
    return result


def read_actions_over_time(filename: str):
    with open(filename) as f:
        data = json.load(f)
    actions_over_time = []
    for key, value in data["Actions"].items():
        parsed_time = dt.datetime.fromisoformat(key.rstrip("Z"))
        actions_over_time.append((parsed_time, value))
    return sorted(actions_over_time, key=lambda a: a[0])
