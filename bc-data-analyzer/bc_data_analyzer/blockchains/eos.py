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
            category["name"]: i
            for i, category in enumerate(self.categories["categories"])
        }
        self.accounts = json.loads(
            pkgutil.get_data(settings.PACKAGE_NAME, "data/eos-accounts.json")
        )

    def compute_categories(self, actions_count):
        category_counts = [0 for _ in self.categories["categories"]]
        for action in actions_count["Actions"]:
            category = self.categories["mapping"].get(action["Name"], "others")
            category_counts[self.category_indexes[category]] += action["Count"]
        return category_counts

    def transform_actions_over_time(
        self, actions: List[Tuple[dt.datetime, dict]]
    ) -> Tuple[List[dt.datetime], List[str], np.ndarray, List[str]]:
        labels = [a["name"].capitalize()
                  for a in self.categories["categories"]]
        dates = [a[0] for a in actions]
        ys = zip(*[self.compute_categories(a) for _, a in actions])
        colors = plot_utils.make_palette(
            *[a["color"] for a in self.categories["categories"]]
        )
        return labels, dates, ys, colors

    @property
    def available_tables(self):
        return ["top-actions"]

    def _generate_table(self, table_name: str, data: dict) -> str:
        if table_name == "top-actions":
            return self._output_top_actions_table(data["Results"]["ActionsByReceiver"])
        raise ValueError("unknown table type {0}".format(table_name))

    def _output_top_actions_table(self, actions: dict) -> str:
        def format_account(account_actions):
            total_actions_count = account_actions["Names"]["TotalCount"]
            actions = [
                a
                for a in account_actions["Names"]["Actions"][:3]
                if a["Count"] / total_actions_count > 0.1
            ]

            def multirow(text, n=len(actions)):
                if n <= 1:
                    return text
                else:
                    return f"\\multirow{{{n}}}{{*}}{{{text}}}"

            def make_action(action):
                name = action["Name"]
                percentage = action["Count"] / total_actions_count * 100
                return f"{name} & {percentage:.2f}\\%"

            name = account_actions["Name"]
            description = self.accounts.get(name, "Unknown")
            formatted_total = f"{total_actions_count:,}"
            rows = [
                f"{multirow(name)} & {multirow(description)} & "
                f"{multirow(formatted_total)} & {make_action(actions[0])}"
            ]
            for action in actions[1:]:
                rows.append(f"& & & {make_action(action)}")
            return "\\\\\n".join(rows)

        rows = [format_account(account) for account in actions["Actions"][:5]]
        return "\\\\\n\\midrule\n".join(rows)
