This directory contains the core functionality of our schlubbin' game.

The important directories from an architectural perspective are:

  * game
    * contains scaffolding for starting up a client as well as a server and connecting them together. This is where we'd define the use of scenes, menus, etc.
  * client
    * contains code for all client behavior/functionality -- uses Ebitengine
  * server
    * contains code for running a game and sending data to clients.
  * world
    * contains shared data used by the client and server. In general most things in world are fully self-contained, with some exceptions for potentially using messages or logging.

Less important directories are:

  * log
    * contains helper functions for logging.
  * transitions
    * fade in/out etc
  * message
    * Our networking messages/events.