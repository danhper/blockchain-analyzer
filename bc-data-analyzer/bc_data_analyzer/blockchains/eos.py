import datetime as dt
from typing import List, Tuple
import json
import pkgutil

import numpy as np

from bc_data_analyzer import plot_utils
from bc_data_analyzer import settings
from bc_data_analyzer.blockchain import Blockchain


@Blockchain.register("eos")
class EOS(Blockchain):
    def __init__(self):
        categories = pkgutil.get_data(
            settings.PACKAGE_NAME, "data/eos-categories.json")
        self.categories = json.loads(categories)
        self.category_indexes = {
            category: i for i, category in enumerate(self.categories["names"])}

    def compute_categories(self, actions_count):
        category_counts = [0 for _ in self.categories["names"]]
        for action in actions_count["Actions"]:
            category = self.categories["mapping"].get(action["Name"], "others")
            category_counts[self.category_indexes[category]] += action["Count"]
        return category_counts

    def transform_actions_over_time(self, actions: List[Tuple[dt.datetime, dict]]) \
            -> Tuple[List[dt.datetime], List[str], np.ndarray, List[str]]:

        labels = [a.capitalize() for a in self.categories["names"]]
        dates = [a[0] for a in actions]
        ys = zip(*[self.compute_categories(a) for _, a in actions])
        colors = plot_utils.make_palette("green", "gray", "teal", "violet", "blue", "brown")
        return labels, dates, ys, colors
