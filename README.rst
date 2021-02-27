***
JAJ
***

Improve YAY
###########

Agenda
======

Goals
-----

1. Improve maintainability by
2. Adding tests and
3. Fixing bugs

Steps
-----

1. Reduce dependencies/coupling
2. Improve testability
3. Add tests
4. Larger architectural changes

Step 1
^^^^^^

* Reduce intra module dependencies

  completion -> db (not possible, but use a smaller interface) [done]

  settings -> vcs [done]

  intrange -> stringset (ParseNumberMenu should be moved) [done]

* Reduce extra module dependencies

  pkg/dep -> aur (via query) [done]

  most dependencies on alpm (via type alias in db) [done]

  most dependencies on gotext (move functionality into pkg/text) [done]

  pkg/upgrade -> alpm (add VerComp method to db) [done]

  pkg/errors -> errors

* Reduce standard library dependencies

  \* -> os and fmt (make pkg/text the destination for all user relevant output) [done]

  news and completion -> net/http (create a http client in main and inject it)


leaky abstractions
^^^^^^^^^^^^^^^^^^

main

  install.go: alpm.QuestionType and bit mask (move to pkg/db)

  install.go: go-srcinfo Parse file (maybe move to pgp)

  query.go: alpm.PkgReasonExplicit (move to pkg/db)

  query.go: rpc.SearchBy (move to pkg/query), what about ``By``?

  cmd.go: alpm.Version

settings

  parser.go: rpc.AURURL [done] (moved to main)

Split parsing off of settings [done]

Move settings/exe a level up (no cohesion with settings) [done]

Investigate merging putting Arguments inside Runtime [see below]
  settings will export three data structures

  1. yay configuration relevant options
  2. pacman configuration relevant options and an easy way to manipulate them
  3. misc options that are only of immediate relevancy [embeds yay config]

More emphasis on pkg/settings/exec. Needs to play well with pkg/text

Make Runtime hold all high level dependencies [done]

Static data structures for pacman options [done]

Removed globals form settings package [done] (moved into db.Executor interface)

Move Runtime into package main

go-srcinfo type alias in pgp/vcs?

research argument handling of '-' (non-blocking read and reattach to console)

::

    db: -
    intrange: -
    multierror: -
    stringset: -
    text: -

    completion: db, text
    exe: text

    vcs: exe, text

    settings: exe, text, vcs

    query: intrange, multierror, text, stringset, db, settings
    news: settings, text

    dep: query, text, db, stringset, settings
    upgrade: db, query, text, vcs, intrange

    pgp: dep, text

    main: *

Now::

    db: -
    intrange: -
    multierror: -
    stringset: -
    text: -
    
    completion: text db
    settings: text
    exe: text
    
    news: text settings
    query: intrange multierror stringset text settings db
    vcs: text exe
    
    dep: stringset text query settings db
    runtime: vcs exe settings db
    upgrade: intrange text vcs query db
    
    db/ialpm: text upgrade settings db
    pgp: text dep

    main: *
    