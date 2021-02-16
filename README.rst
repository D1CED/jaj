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

  completion -> db (not possible, but use a smaller interface)

  dep -> settings

  query -> settings

  news -> settings

  intrange -> stringset (ParseNumberMenu should be moved)

* Reduce extra module dependencies

  pkg/dep -> aur (via query)

  most dependencies on alpm (via type alias in db) [done]

  most dependencies on gotext (move functionality into pkg/text)

  pkg/upgrade -> alpm (add VerComp method to db)

* Reduce standard library dependencies

  \* -> os and fmt (make pkg/text the destination for all user relevant output)

  news and completions -> net/http (create a http client in main and inject it)
