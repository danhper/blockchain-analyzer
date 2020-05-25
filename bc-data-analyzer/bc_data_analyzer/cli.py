import argparse

from bc_data_analyzer import commands


parser = argparse.ArgumentParser(prog="data-analyzer")
parser.add_argument("-b", "--blockchain", required=True,
                    help="Which blockchain to use")

subparsers = parser.add_subparsers(dest="command")

plot_action_over_time = subparsers.add_parser("plot-actions-over-time", help="Plot actions over time")
plot_action_over_time.add_argument("input", help="Input file containing actions over time")
plot_action_over_time.add_argument("-o", "--output", help="Output file")


def run():
    args = vars(parser.parse_args())
    if not args["command"]:
        parser.error("no command given")
    
    func = getattr(commands, args["command"].replace("-", "_"))
    func(args)
