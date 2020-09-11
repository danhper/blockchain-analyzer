import bisect
import datetime as dt

import numpy as np
import matplotlib.pyplot as plt
import matplotlib.dates as mdates
import matplotlib

from bc_data_analyzer import settings


COLORS = {
    "green": "olivedrab",
    "blue": "cornflowerblue",
    "brown": "darkgoldenrod",
    "gray": "darkgray",
    "violet": "violet",
    "coral": "lightcoral",
    "teal": "teal",
}


def make_palette(*colors):
    return [COLORS[c] for c in colors]


def show_plot(filename):
    if filename is None:
        plt.show()
    else:
        plt.savefig(filename)


def find_range(dates):
    start_date, end_date = settings.START_DATE, settings.END_DATE
    if dates[0].tzinfo is None:
        start_date, end_date = start_date.replace(tzinfo=None), end_date.replace(tzinfo=None)
    start_index = bisect.bisect_right(dates, start_date) - 1
    end_index = bisect.bisect_left(dates, end_date)
    return max(start_index, 0), end_index


def adjust_series(x, ys):
    if len(x) and isinstance(x[0], np.datetime64):
        x = [dt.datetime.utcfromtimestamp(v / 1e9) for v in x.tolist()]
    start_index, end_index = find_range(x)
    x = x[start_index:end_index]
    ys = [y[start_index:end_index] for y in ys]
    return x, ys


def plot_chart_area(labels, x, *ys, filename=None, **kwargs):
    matplotlib.rc("font", size=14)
    x, ys = adjust_series(x, ys)

    fig, ax = plt.subplots(figsize=(10, 7))
    if "ylim" in kwargs:
        plt.ylim(top=kwargs.pop("ylim"))
    plt.xticks(rotation=45)
    plt.setp(ax.xaxis.get_majorticklabels(), ha="right")
    ax.set_ylabel("Number of Actions")
    ax.stackplot(x, *ys, labels=labels, **kwargs)
    ax.xaxis.set_major_formatter(mdates.DateFormatter("%Y-%m-%d"))
    ax.ticklabel_format(scilimits=(0, 0), axis="y")
    plt.legend(loc="upper left")
    plt.tight_layout()
    show_plot(filename)


def plot_transaction_volume(x, y_tx_count, y_amount_usd, filename=None):
    x, (y_tx_count, y_amount_usd) = adjust_series(x, [y_tx_count, y_amount_usd])

    _fig, ax = plt.subplots()
    plt.xticks(rotation=45)
    plt.setp(ax.xaxis.get_majorticklabels(), ha="right")
    ax.plot(x, y_tx_count, color=COLORS["blue"])
    ax2 = ax.twinx()
    ax2.plot(x, y_amount_usd, color=COLORS["green"])
    ax.figure.legend(["Transactions count", "USD volume"],
                     bbox_to_anchor=(1.0, 0.1), frameon=False)
    ax2.set_yscale("log")
    ax.set_xlabel("Time")
    ax.set_ylabel("Transactions count")
    ax2.set_ylabel("USD volume")
    ax.xaxis.set_major_formatter(mdates.DateFormatter("%m-%d"))
    plt.tight_layout()
    if filename:
        plt.savefig(filename)
    else:
        plt.show()
