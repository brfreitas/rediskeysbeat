from rediskeysbeat import BaseTest

import os


class Test(BaseTest):

    def test_base(self):
        """
        Basic test with exiting Rediskeysbeat normally
        """
        self.render_config_template(
            path=os.path.abspath(self.working_dir) + "/log/*"
        )

        rediskeysbeat_proc = self.start_beat()
        self.wait_until(lambda: self.log_contains("rediskeysbeat is running"))
        exit_code = rediskeysbeat_proc.kill_and_wait()
        assert exit_code == 0
