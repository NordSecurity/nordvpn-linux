import random
from functools import wraps

import pytest
from itertools import product


def _is_list_of_lists(obj):
    """
    Check if the object is a list/tuple of lists/tuples.

    :param obj: Object to check.
    :return: True if obj is a list/tuple of lists/tuples, else False.
    """
    return bool(isinstance(obj, list | tuple) and obj and all(isinstance(x, list | tuple) for x in obj))


def _flatten_param_tuple(combo):
    """
    Flatten any nested tuples/lists inside a combo into a single tuple.

    :param combo: An iterable of parameters.
    :return: A flattened tuple of values.
    """
    flat = []
    for item in combo:
        if isinstance(item, tuple | list):
            flat.extend(item)
        else:
            flat.append(item)
    return tuple(flat)


def as_tuple(obj):
    """
    Converts the input to a tuple.

    If the input is already a tuple or a list, it is converted to a tuple (if not already).
    If the input is any other type, it is returned as a single-element tuple containing the input.

    :param obj: An iterable of parameters. The input object to be converted.
    :return: A tuple representation of the input.
    """
    return tuple(obj) if isinstance(obj, list | tuple) else (obj,)


def _generate_params_with_ids(argnames, randomized_source, ordered_source, generate_all=False, id_pattern=None, always_pair=None, sample_size=None):
    """
    Generate test parameters and their associated unique identifiers (IDs).

    This function creates a combination of parameters from two data sources:
    randomized data and ordered data. Additionally, test case IDs
    are constructed to provide human-readable descriptions for test output.

    :param argnames: A list of argument names for the test function.
    :param randomized_source:
        A list of items that will be used as randomized data.
    :param ordered_source:
        A list of ordered items.
    :param generate_all: If True, generate the full cross product (cartesian product) of all sources.
    :param id_pattern:
        A format string for generating unique test IDs.
        - {randomized} is replaced by the corresponding value from the randomized source.
        - {ordered} is replaced by the corresponding value from the ordered source.
    :param always_pair: item from randomized_source to always be paired with every ordered value
    :param sample_size: Randomly select this many cases from all cartesian combinations (cartesian mode only).

    :return:
        - params: A list of parameter combining randomized and ordered data.
        - ids: A list of human-readable strings with names for each test case.
    """
    params = []
    ids = []

    if generate_all:
        rnds = randomized_source if _is_list_of_lists(randomized_source) else [randomized_source]
        ords = ordered_source if _is_list_of_lists(ordered_source) else [ordered_source]
        sources = list(ords) + list(rnds)
        all_combos = [_flatten_param_tuple(combo) for combo in product(*sources)]
        if sample_size is not None:
            all_combos = random.Random().sample(all_combos, min(sample_size, len(all_combos)))
    else:
        rnds = _flatten_param_tuple(randomized_source) if _is_list_of_lists(randomized_source) else [randomized_source]
        ords = _flatten_param_tuple(ordered_source) if _is_list_of_lists(ordered_source) else [ordered_source]
        all_combos = []

        # Always pair the exact always_pair item (if given) for each ordered_value
        if always_pair:
            for ord_elem in ords:
                all_combos.append(as_tuple(ord_elem) + as_tuple(always_pair))
            rnds = [item for item in rnds if item != always_pair]

        for ord_elem in ords:
            rand_elem = random.Random().choice(rnds)
            combo = as_tuple(ord_elem) + as_tuple(rand_elem)
            all_combos.append(combo)

    for flat_combo in all_combos:
        params.append(flat_combo)
        if id_pattern:
            arg_dict = dict(zip(argnames, flat_combo, strict=False))
            ids.append(id_pattern.format(**arg_dict))
        else:
            ids.append("".join(str(v) for v in flat_combo))
    return params, ids


def dynamic_parametrize(argnames, randomized_source, ordered_source, generate_all=False, id_pattern=None, always_pair=None, sample_size=None):
    """
    A pytest decorator for dynamic parameterization of test functions.

    This decorator applies dynamic parameters ({randomized_source} + {ordered_source}) to pytest test functions,
    automating the generation of parameterized test cases while supporting human-readable test IDs.

    :param argnames:
        A string or list of strings specifying the argument names for the parameterized test function
    :param randomized_source:
        A list of items that will be used as randomized data.
    :param ordered_source:
        A list of ordered items.
    :param generate_all: If True, generate the full cross product (cartesian product) of all sources.
    :param id_pattern:
        A format string for generating unique test IDs.
        - {randomized_source} is replaced by the corresponding value from the randomized source.
        - {ordered_source} is replaced by the corresponding value from the ordered source.
    :param always_pair: item from randomized_source to always be paired with every ordered value
    :param sample_size: Randomly select this many cases from all cartesian combinations (cartesian mode only).

    :return:
        The dynamically parameterized test function wrapped with pytest's pytest.mark.parametrize.
    """
    def decorator(func):
        params, ids = _generate_params_with_ids(argnames, randomized_source, ordered_source, generate_all,
                                                id_pattern, always_pair, sample_size)

        @pytest.mark.parametrize(argnames, params, ids=ids)
        @wraps(func)
        def wrapper(*args, **kwargs):
            return func(*args, **kwargs)

        return wrapper

    return decorator
