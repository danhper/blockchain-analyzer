import json

from bc_data_analyzer.blockchain import Blockchain
from bc_data_analyzer import data_reader
from bc_data_analyzer import plot_utils


def plot_actions_over_time(args):
    actions_over_time = data_reader.read_actions_over_time(args["input"])
    blockchain = Blockchain.create(args["blockchain"])
    labels, dates, ys, colors = blockchain.transform_actions_over_time(actions_over_time)
    plot_utils.plot_chart_area(labels, dates, *ys, colors=colors, filename=args["output"])


def generate_table(args):
    with open(args["input"]) as f:
        data = json.load(f)
    blockchain: Blockchain = Blockchain.create(args["blockchain"])
    table = blockchain.generate_table(args["name"], data)
    print(table)
