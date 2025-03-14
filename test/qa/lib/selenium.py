import os

import sh
from selenium import webdriver
from selenium.webdriver.common.by import By
from selenium.webdriver.support.expected_conditions import presence_of_element_located
from selenium.webdriver.support.ui import WebDriverWait

# Path where we save screenshots from browser, of Selenium tests, if they fail
BROWSER_LOGS_PATH = f"{os.environ['WORKDIR']}/dist/logs/" #browser_tests/"

LOGIN_MSG_SUCCESS = "Welcome to NordVPN! You can now connect to VPN by using 'nordvpn connect'."
LOGOUT_MSG_SUCCESS = "You are logged out."

# environment variable, which we will use to specify location to temporary folder for Firefox
ENV_FF_TMP = "TMPDIR"

# location to store temporary Firefox files, in case Firefox was installed from Snap (prevents profile error)
SNAP_FF_TMP_DIR_PATH = os.path.expanduser("~") + "/ff_tmp"

# XPath used in Selenium tests, to determine locations of elements on NordAccount
NA_USERNAME_PAGE_TEXTBOX_XPATH = '//*[@data-testid="identifier-input"]'
NA_USERNAME_PAGE_BUTTON_XPATH = '//*[@data-testid="identifier-submit"]'

NA_PASSWORD_PAGE_TEXTBOX_XPATH = '//*[@data-testid="signin-password-input"]'  # noqa: S105
NA_PASSWORD_PAGE_BUTTON_XPATH = '//*[@data-testid="signin-button"]'  # noqa: S105

NA_CONTINUE_PAGE_LINK_BUTTON = '//*[@data-testid="error-page-link"]'

# used with Selenium login tests
LOGIN_FLAG = ["", "--nordaccount"]


class SeleniumBrowser:
    def __init__(self, preferences:list | None=None):
        self.options = webdriver.FirefoxOptions()
        self.options.binary_location = str(sh.which("firefox"))
        self.options.add_argument('--headless')
        self.options.add_argument('--no-sandbox')

        if preferences is not None:
            for preference, value in preferences:
                self.options.set_preference(preference, value)

        if os.path.exists("/snap/firefox") and os.environ.get(ENV_FF_TMP) is None:
            os.mkdir(SNAP_FF_TMP_DIR_PATH)
            os.environ[ENV_FF_TMP] = SNAP_FF_TMP_DIR_PATH

        service = webdriver.FirefoxService(executable_path="/usr/bin/geckodriver") #, log_output=BROWSER_LOGS_PATH + "geckodriver.log")
        self.browser = webdriver.Firefox(options=self.options, service=service)

    def browser_get(self) -> webdriver.Firefox:
        return self.browser

    def browser_kill(self) -> None:
        """Quits browser, and deletes temporary folder created for Firefox."""
        self.browser.quit()

        # Cleanup, if we were working with Firefox from snap
        if os.environ.get(ENV_FF_TMP) is not None:
            os.environ.pop(ENV_FF_TMP)
            os.removedirs(SNAP_FF_TMP_DIR_PATH)

    def browser_element_interact(self, xpath: str, write:str | None=None, return_attribute:str | None=None) -> None | str:
        """
        Clicks element on website, specified by `xpath`.

        If `write` parameter value is set, also writes the set value to the element.

        If `return_attribute` parameter value is set, also returns specified attribute value of element.
        """

        # Defines how long will we wait for elements, that we are looking for to appear in pages
        wait = WebDriverWait(self.browser, 10)

        website_element = wait.until(presence_of_element_located((By.XPATH, xpath)))
        website_element.click()

        if write is not None:
            website_element.send_keys(write)

        if return_attribute is not None:
            return website_element.get_attribute(return_attribute)
        return None
