from abc import ABC, abstractmethod
import datetime as dt
from typing import List, Tuple

import numpy as np

from bc_data_analyzer.base_factory import BaseFactory


class Blockchain(BaseFactory):
    @abstractmethod
    def transform_actions_over_time(self, actions: List[Tuple[dt.datetime, dict]]) \
        -> Tuple[List[dt.datetime], List[str], List[np.ndarray]]:
        pass

    @abstractmethod
    def generate_table(self, table_name: str, data: dict) -> str:
        pass
