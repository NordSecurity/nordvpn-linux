import random
from functools import wraps

import pytest


def _generate_params_with_ids(randomized_source, ordered_source, id_pattern=None):
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
        - `{0}` refers to the first element of randomized data, `{1}` to the second, etc.
        - `{ordered}` is replaced by the corresponding value from the ordered source.

    :return:
        - `params`: A list of parameter combining randomized and ordered data.
        - `ids`: A list of human-readable strings with names for each test case.
    """
    params = []
    ids = []

    for ordered_value in ordered_source:
        randomized_value = random.choice(randomized_source)
        param_tuple = (*randomized_value, ordered_value)
        params.append(param_tuple)

        if id_pattern:
            ids.append(id_pattern.format(*randomized_value, ordered=ordered_value))
        else:
            ids.append(f"ordered={ordered_value}, randomized={randomized_value}")

    return params, ids


def dynamic_parametrize(argnames, randomized_source, ordered_source, id_pattern=None):
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
        - `{0}` refers to the first element of randomized data, `{1}` to the second, etc.
        - `{ordered}` is replaced by the corresponding value from the ordered source.

    :return: Callable
        The dynamically parameterized test function wrapped with pytest's `pytest.mark.parametrize`.
    """
    def decorator(func):

        params, ids = _generate_params_with_ids(randomized_source, ordered_source, id_pattern)

        @pytest.mark.parametrize(argnames, params, ids=ids)
        @wraps(func)
        def wrapper(*args, **kwargs):
            return func(*args, **kwargs)

        return wrapper

    return decorator
