.. _test:

================
Testing Binaries
================

.. contents:: :local:
   :depth: 1

Xplugeth includes a functional test to verify the correctness of a built binary by running a local node on Holesky, capturing data, and comparing it against pre-collected blockchain data.

Prerequisites
-------------
Ensure you install the required dependencies from the test suite before running the test:

.. code-block:: shell

    pip install -r test/resources/requirements.txt


Running the Test
----------------
To execute the test, you need a **compiled xplugeth binary**. The test requires the binary path to be provided. To get a binary refer to `this section <build.html#example-building-a-binary>`_.

**Run with pytest:**

.. code-block:: shell

   pytest test/test_main.py --bin-path=/path/to/binary

**Run without pytest (Debug Mode):**

.. code-block:: shell

   python3 test/test_main.py /path/to/binary

Replace ``/path/to/xplugeth`` with the actual path of the built binary.

If the test completes successfully, it means the binary behaves as expected. Otherwise, pytest will highlight any discrepancies.
