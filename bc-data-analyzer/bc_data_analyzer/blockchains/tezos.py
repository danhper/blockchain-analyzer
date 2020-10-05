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

    def transform_actions_over_time(
        self, actions: List[Tuple[dt.datetime, dict]]
    ) -> Tuple[List[dt.datetime], List[str], np.ndarray, List[str]]:
        actions_count = count_actions_over_time(actions)
        sorted_actions = sorted(actions_count.items(), key=lambda v: -v[1])
        top_actions = [k for k, _ in sorted_actions[: self.max_actions]]

        labels = [a.capitalize() for a in top_actions] + ["Other"]
        dates = [a[0] for a in actions]
        ys = np.array(
            [self._transform_action(a["Actions"], top_actions)
             for _, a in actions]
        ).T
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
        result.append(sum(a["Count"]
                          for a in actions if a["Name"] not in top_actions))
        return result

    @property
    def available_tables(self):
        return ["top-senders"]

    def _generate_table(self, table_name: str, data: dict) -> str:
        if table_name == "top-senders":
            return self._output_top_senders_table(data["Results"]["ActionsBySender"])
        raise ValueError("unknown table type {0}".format(table_name))

    def _output_top_senders_table(self, data: dict) -> str:
        def make_row(row):
            receivers_count = row["Receivers"]["UniqueCount"]
            count = row["Count"]
            row_data = dict(
                name=row["Name"],
                count=count,
                avg=count / receivers_count,
                unique_count=row["Receivers"]["UniqueCount"],
            )
            return (
                r"\tezaddr{{{name}}} & {count:,} & {unique_count:,} & {avg:.2f}".format(
                    **row_data
                )
            )

        rows = "\\\\\n    ".join([make_row(row)
                                  for row in data["Actions"][:5]])
        return r"""\begin{{figure*}}[tbp]
    \footnotesize
    \centering
    \begin{{tabular}}{{l r r r}}
    \toprule
               &                &               & \bf Avg. \#\\
               &                & \bf Unique    & \bf of transactions\\
    \bf Sender & \bf Sent count & \bf receivers & \bf per receiver\\
    \midrule
    {rows}\\
    \bottomrule
    \end{{tabular}}
    \caption{{Tezos accounts with the highest number of sent transactions.}}
    \label{{tab:tezos-account-edges}}
\end{{figure*}}""".format(
            rows=rows
        )
