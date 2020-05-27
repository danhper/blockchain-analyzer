from collections import defaultdict
import datetime as dt
from typing import List, Tuple, Dict


def count_actions_over_time(actions: List[Tuple[dt.datetime, dict]]) -> Dict[str, int]:
    result = defaultdict(int)
    for _, actions_count in actions:
        for action in actions_count["Actions"]:
            result[action["Name"]] += action["Count"]
    return result
