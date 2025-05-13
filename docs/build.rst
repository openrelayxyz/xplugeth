.. _build:

================
Using Xplugeth
================

.. contents:: :local:
   :depth: 1

Setting Up 
==========
Before using the build tool to generate binaries, ensure your system has the required dependencies:

- Python 3.12.3 or later
- Go 1.22 or later
- Git

To install dependencies, run:

.. code-block:: shell

   pip install -r build/requirements.txt

Using the Build Tool
====================

To build a binary, run:

.. code-block:: shell

   python3 build/build.py --workdir /tmp/build_xplugeth

This will:

- Clone the Geth repository 
- Checkout Geth version (if specified)
- Apply patches  
- Integrate plugins (if specified)  
- Compile the final binary  

By default, the output binary is stored in ``/tmp/output/``. 

The build tool provides several options for customization. If you run with the ``--help`` flag, it shows you all the available arguments that can be passed in with flags

.. code-block:: shell

   python3 build/build.py --help

Flags
-----

- ``--source-remote``, ``-s``  
  This flag specifies the Git repository to clone for the build.  
  **Default:** https://github.com/ethereum/go-ethereum

- ``--source-tag``, ``-t``  
  Specifies the version or tag of Geth to build.  
  **Default:** ``v1.14.13``

- ``--plugin``, ``-p``  
  Adds one or more plugins to the build, provided as import paths.  
  **Example:**  

  .. code-block:: shell

     --plugin github.com/openrelayxyz/xplugeth/plugins/merge@v0.12.0 \
     --plugin github.com/openrelayxyz/xplugeth/plugins/producer@v0.12.0

- ``--replace``, ``-r``  
  Replaces a package dependency with a local version.

  .. code-block:: shell

      --replace github.com/example/module=../local/module

- ``--cmd``, ``-c``  
  Specifies the command directory where the build process starts. Default: ``./cmd/geth``

- ``--workdir``, ``-w``  
  Defines the working directory for cloning and patching.  

- ``--artifacts-directory``, ``-a``  
  Sets the output directory for the compiled binary.  
  **Default:** ``/tmp/output/``.

- ``--archive``, ``-v``  
  Specifies a Git archive repository where the built binary can be pushed. If no value is provided, it defaults to  
  ``git@github.com:openrelayxyz/xplugeth-archive.git``.  
  **Note:** The archive URL must be an SSH URL to preserve the user's Git credentials.  
  **Example:**  

  .. code-block:: shell

     --archive git@github.com:example/xplugeth-archive.git



.. _example-building-a-binary:

Example: Building a Binary
==========================

- To build a binary with multiple plugins, a specific Geth version, and custom directories, run the code below. Please replace ``/workdir`` and ``artifactsdir`` with actual paths where the code and binary should live.

   .. code-block:: shell

        python3 build/build.py \
            -s https://github.com/ethereum/go-ethereum/ \
            -t v1.14.13 \
            -p github.com/openrelayxyz/xplugeth/plugins/merge@v0.12.0 \
            -p github.com/openrelayxyz/xplugeth/plugins/producer@v0.12.0 \
            -c ./cmd/cli \
            -w /workdir \
            -a /artifactsdir



- To build a vanilla geth binary without any plugin, you can simply run the file without passing any arguments as it assumes the defaults.  
   
   .. code-block:: shell

      python3 build/build.py 
