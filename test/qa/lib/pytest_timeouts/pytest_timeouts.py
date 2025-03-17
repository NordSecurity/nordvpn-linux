from __future__ import absolute_import # noqa: UP010

import functools
import signal

import pytest

SETUP_TIMEOUT_HELP = 'test case setup timeout in seconds'
EXECUTION_TIMEOUT_HELP = 'test case execution timeout in seconds'
TEARDOWN_TIMEOUT_HELP = 'test case teardown timeout in seconds'
TIMEOUT_ORDER_HELP = """override order: i - ini, m - mark, o - opt
example: "omi", "imo", "i" - ini only
"""


@staticmethod
def get_markers_old_way(item, name):
    return item.get_marker(name=name)


@staticmethod
def get_markers_new_way(item, name):
    return item.iter_markers(name=name)


@pytest.hookimpl
def pytest_addoption(parser):
    group = parser.getgroup('timeouts')
    group.addoption(
        '--setup-timeout',
        type=float,
        help=SETUP_TIMEOUT_HELP,
    )
    group.addoption(
        '--execution-timeout',
        type=float,
        help=EXECUTION_TIMEOUT_HELP,
    )
    group.addoption(
        '--teardown-timeout',
        type=float,
        help=TEARDOWN_TIMEOUT_HELP,
    )
    group.addoption(
        '--timeouts-order',
        type=str,
        help=TIMEOUT_ORDER_HELP,
        default='omi'
    )
    parser.addini('setup_timeout', SETUP_TIMEOUT_HELP)
    parser.addini('execution_timeout', SETUP_TIMEOUT_HELP)
    parser.addini('teardown_timeout', SETUP_TIMEOUT_HELP)


@pytest.hookimpl
def pytest_configure(config):
    assert hasattr(signal, 'SIGALRM')
    TimeoutsPlugin.configure()
    config.pluginmanager.register(TimeoutsPlugin(config))


class TimeoutsPlugin(object): # noqa: UP004
    def __init__(self, config):
        config.addinivalue_line(
            'markers',
            'execution_timeout(seconds): '
            'time out test case after specified time\n'
        )
        config.addinivalue_line(
            'markers',
            'setup_timeout(seconds): '
            'time out fixture setup after specific time\n'
        )
        config.addinivalue_line(
            'markers',
            'teardown_timeout(seconds):'
            'time out fixture teardown after specific time\n'
        )
        self.order = self.fetch_timeout_order(config)
        self.timeout = {
            'setup_timeout': self.fetch_timeout_from_config(
                'setup_timeout', config),
            'execution_timeout': self.fetch_timeout_from_config(
                'execution_timeout', config),
            'teardown_timeout': self.fetch_timeout_from_config(
                'teardown_timeout', config),
        }

    @staticmethod
    def parse_timeout(timeout):
        timeout = (
            0.0 if (timeout is None) or (timeout == '')
            else float(timeout)
        )
        timeout = 0.0 if timeout < 0.0 else timeout
        return timeout

    @staticmethod
    def configure():
        ver = [int(v) for v in pytest.__version__.split('.')]
        if (ver[0] > 3) or ((ver[0] == 3) and (ver[1] >= 6)):
            TimeoutsPlugin.get_markers = get_markers_new_way
        else:
            TimeoutsPlugin.get_markers = get_markers_old_way

    @staticmethod
    def fetch_timeout_from_config(timeout_name, config):
        timeout_option = config.getvalue(timeout_name)
        timeout_ini = config.getini(timeout_name)
        return timeout_option, timeout_ini

    @staticmethod
    def fetch_timeout_order(config):
        order = list(config.getvalue('timeouts_order'))
        order_set = set(['i', 'm', 'o'])
        if len(order) == 0 or len(order) > 3:
            raise pytest.UsageError(
                'Order should have at least 1 and less then or '
                'equal 3 elements'
            )
        if not set(order).issubset(order_set):
            raise pytest.UsageError(
                'Incorrect item \'{}\' in timeout order list'.format( # noqa: UP032
                    list(set(order).difference(order_set)))
            )
        return order

    def fetch_timeout(self, timeout_name, item):
        marker_timeout = (
            self.fetch_marker_timeout(item, timeout_name) if item is not None
            else None
        )
        timeout = None
        for order_item in self.order:
            if order_item == 'o' and self.timeout[timeout_name][0] is not None:
                timeout = self.timeout[timeout_name][0]
                break
            elif order_item == 'm' and marker_timeout is not None: # noqa: RET508
                timeout = marker_timeout
                break
            elif (order_item == 'i' and
                  self.timeout[timeout_name][1] != ''):
                timeout = self.timeout[timeout_name][1]
                break
        return self.parse_timeout(timeout)

    @pytest.hookimpl(tryfirst=True)
    def pytest_report_header(self, config):
        timeout_prints = [
            'setup timeout: {}s'.format(
                self.fetch_timeout('setup_timeout', None)),
            'execution timeout: {}s'.format(
                self.fetch_timeout('execution_timeout', None)),
            'teardown timeout: {}s'.format(
                self.fetch_timeout('teardown_timeout', None)),
        ]
        return [', '.join(timeout_prints)]

    @pytest.hookimpl
    def pytest_enter_pdb(self):
        self.cancel_timer()

    @pytest.hookimpl(hookwrapper=True)
    def pytest_runtest_setup(self, item):
        self.setup_timer(self.fetch_timeout('setup_timeout', item))
        yield
        self.cancel_timer()

    @pytest.hookimpl(hookwrapper=True)
    def pytest_runtest_call(self, item):
        self.setup_timer(self.fetch_timeout('execution_timeout', item))
        yield
        self.cancel_timer()

    @staticmethod
    def fetch_marker_timeout(item, name):
        def get_fixture_scope(item):
            return item._fixtureinfo.name2fixturedefs[
                item._fixtureinfo.names_closure[0]][0].scope
        markers = TimeoutsPlugin.get_markers(item, name)
        if markers:
            for marker in markers:
                if marker.args:
                    if len(marker.args) == 2:
                        if marker.args[1] == get_fixture_scope(item):
                            return marker.args[0]
                        else: # noqa: RET505
                            continue
                    else:
                        return marker.args[0]
                else:
                    raise TypeError('Timeout value is missing')
        return None

    @pytest.hookimpl(hookwrapper=True)
    def pytest_runtest_teardown(self, item):
        self.setup_timer(self.fetch_timeout('teardown_timeout', item))
        yield
        self.cancel_timer()

    @staticmethod
    def setup_timer(timeout):
        handler = functools.partial(TimeoutsPlugin.timeout_handler, timeout)
        signal.signal(signal.SIGALRM, handler)
        signal.setitimer(signal.ITIMER_REAL, timeout)

    @staticmethod
    def cancel_timer():
        signal.setitimer(signal.ITIMER_REAL, 0)
        signal.signal(signal.SIGALRM, signal.SIG_DFL)

    @staticmethod
    def timeout_handler(timeout, signum, frame):
        __tracebackhide__ = True
        pytest.fail('Timeout >%ss' % timeout) # noqa: UP031
