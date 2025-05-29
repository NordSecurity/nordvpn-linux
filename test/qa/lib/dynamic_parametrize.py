import random
from functools import wraps

import pytest


def _generate_params_with_ids(randomized_source, ordered_source, id_pattern=None, always_pair=None):
    """
    Generate test parameters and their associated unique identifiers (IDs).

    This function creates a combination of parameters from two data sources:
    randomized data and ordered data. Additionally, test case IDs
    are constructed to provide human-readable descriptions for test output.

    :param randomized_source:
        A list of items that will be used as randomized data.
    :param ordered_source:
        A list of ordered items.
    :param id_pattern:
        A format string for generating unique test IDs.
        - `{randomized}` is replaced by the corresponding value from the randomized source.
        - `{ordered}` is replaced by the corresponding value from the ordered source.
    :param always_pair: item from randomized_source to always be paired with every ordered value

    :return:
        - `params`: A list of parameter combining randomized and ordered data.
        - `ids`: A list of human-readable strings with names for each test case.
    """
    params = []
    ids = []

    def add_params_and_ids(randomized, ordered):
        param_tuple = (*randomized, ordered)
        params.append(param_tuple)
        if id_pattern:
            ids.append(id_pattern.format(ordered=ordered, randomized=randomized))
        else:
            ids.append(f"ordered={ordered}, randomized={randomized}")

    # Always pair the exact always_pair item (if given) for each ordered_value
    if always_pair:
        for ordered_value in ordered_source:
            add_params_and_ids(always_pair, ordered_value)
        remaining_randomized_data = [item for item in randomized_source if item != always_pair]
    else:
        remaining_randomized_data = list(randomized_source)

    # For each ordered value, get a random item from remaining_randomized_data
    for ordered_value in ordered_source:
        randomized_value = random.choice(remaining_randomized_data)
        add_params_and_ids(randomized_value, ordered_value)

    return params, ids


def dynamic_parametrize(argnames, randomized_source, ordered_source, id_pattern=None, always_pair=None):
    """
    A pytest decorator for dynamic parameterization of test functions.

    This decorator applies dynamic parameters (`randomized_source` + `ordered_source`) to pytest test functions,
    automating the generation of parameterized test cases while supporting human-readable test IDs.

    :param argnames:
        A string or list of strings specifying the argument names for the parameterized test function
    :param randomized_source:
        A list of items that will be used as randomized data.
    :param ordered_source:
        A list of ordered items.
    :param id_pattern:
        A format string for generating unique test IDs.
        - `{randomized}` is replaced by the corresponding value from the randomized source.
        - `{ordered}` is replaced by the corresponding value from the ordered source.
    :param always_pair: item from randomized_source to always be paired with every ordered value

    :return:
        The dynamically parameterized test function wrapped with pytest's `pytest.mark.parametrize`.
    """
    def decorator(func):

        params, ids = _generate_params_with_ids(randomized_source, ordered_source, id_pattern, always_pair)

        @pytest.mark.parametrize(argnames, params, ids=ids)
        @wraps(func)
        def wrapper(*args, **kwargs):
            return func(*args, **kwargs)

        return wrapper

    return decorator
