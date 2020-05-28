import datetime as dt
from typing import List, Tuple

import numpy as np

from bc_data_analyzer import plot_utils
from bc_data_analyzer.blockchain import Blockchain
from bc_data_analyzer.aggregator import count_actions_over_time


OTHER_KEY = "other"


@Blockchain.register("tezos")
class Tezos(Blockchain):
    def __init__(self, max_actions=2):
        self.max_actions = max_actions

    def transform_actions_over_time(self, actions: List[Tuple[dt.datetime, dict]]) \
            -> Tuple[List[dt.datetime], List[str], np.ndarray, List[str]]:
        actions_count = count_actions_over_time(actions)
        sorted_actions = sorted(actions_count.items(), key=lambda v: -v[1])
        top_actions = [k for k, _ in sorted_actions[:self.max_actions]]

        labels = [a.capitalize() for a in top_actions] + ["Other"]
        dates = [a[0] for a in actions]
        ys = np.array([self._transform_action(a["Actions"], top_actions) for _, a in actions]).T
        return labels, dates, ys, plot_utils.make_palette("blue", "green", "brown")

    @staticmethod
    def _find_action_count(actions: List[dict], name: str) -> int:
        for action in actions:
            if action["Name"] == name:
                return action["Count"]
        return 0

    def _transform_action(self, actions: dict, top_actions: dict) -> dict:
        result = []
        for action in top_actions:
            result.append(self._find_action_count(actions, action))
        result.append(sum(a["Count"] for a in actions if a["Name"] not in top_actions))
        return result