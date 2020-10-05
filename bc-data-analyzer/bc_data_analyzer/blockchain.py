from abc import ABC, abstractmethod
import datetime as dt
from typing import List, Tuple

import numpy as np

from bc_data_analyzer.base_factory import BaseFactory


class Blockchain(BaseFactory, ABC):
    @abstractmethod
    def transform_actions_over_time(self, actions: List[Tuple[dt.datetime, dict]]) \
            -> Tuple[List[dt.datetime], List[str], List[np.ndarray]]:
        pass

    def generate_table(self, table_name: str, data: dict) -> str:
        if table_name not in self.available_tables:
            raise ValueError(
                "unknown table type {0}, available: {1}".format(
                    table_name, ", ".join(self.available_tables)))
        return self._generate_table(table_name, data)

    @abstractmethod
    def _generate_table(self, table_name: str, data: dict) -> str:
        pass

    @property
    @abstractmethod
    def available_tables(self) -> List[str]:
        pass
