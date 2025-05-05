========
Xplugeth
========

Xplugeth is an evolution of `Plugeth <https://github.com/openrelayxyz/plugeth>`_, designed to simplify the process of building customized Geth binaries with integrated modifications, including extended functionality (plugins) and core-level changes (patches).

PluGeth introduced a modular way to extend Geth, but keeping up with upstream changes (Geth, Core-Geth, Matic-Bor) while maintaining compatibility across multiple repositories led to code drift and increased maintenance complexity. Xplugeth addresses this by integrating necessary modifications directly into the codebase, eliminating the need for separate utility repositories and reducing the overhead of maintaining forks like Plugeth-Bor and Plugeth-ETC, ensuring a more streamlined and consistent build process.

Patches
-------

Xplugeth applies patches to modify Geth’s behavior at the source level before compilation. These patches enable **custom functionality** by modifying Geth’s core internals, **adding hooks** for external plugins, and making **network-specific changes** for chains like Polygon and ETC. By using patches, Xplugeth eliminates the need for maintaining multiple forks, ensuring that all modifications are handled dynamically within a unified codebase.

Plugins
-------

Plugins work alongside patches to extend Geth’s capabilities. During the build process, Xplugeth integrates plugins into the binary, allowing them to interact with the patched functionality. This modular approach enables developers to add **custom APIs, transaction handling, or state modifications** without permanently altering Geth’s core code.

- :ref:`build`
- :ref:`test`


.. toctree::
    :maxdepth: 1
    :caption: Overview
    :hidden:
    
    build
    test

.. toctree::
    :maxdepth: 1
    :caption: Contact
    :hidden:

    contact
