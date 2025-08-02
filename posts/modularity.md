---
title: Modularity
date: 2025-08-01
---

Some systems are just too big for anyone to fully understand. So how do we
still build and work on them? The time-tested strategy is *divide and conquer*:
split the system into smaller pieces we can handle, then put them back together
to solve the bigger problem. But splitting alone isn’t enough — the value lies
in *how* we split it, and *how* we connect the parts.

Good modular design is about enabling *local reasoning*. You should be able to
fully understand a module without having to figure out the whole system. You
should be able to change a module without forcing changes everywhere else.
That’s only possible with clear, stable boundaries between modules — in other
words, well-defined *interfaces*. Interfaces aren’t just the glue that connects
modules; they are the abstraction that lets us make a few key assumptions and
safely ignore the rest.

We can think of an interface as a *contract* between suppliers and consumers.
It defines what suppliers require and what they promise in return. A
well-designed interface can also enable *reuse*: if it supports multiple use
cases, we can rely on existing modules instead of building new ones from
scratch. But this flexibility must not come at the cost of local reasoning — an
interface that tries to serve too many purposes often ends up serving none
well.

It's important that we take extra care to specify the contract of each
interface, and to make sure that all modules in the system respect these
contracts. That’s what enables strong and scalable systems. Contracts give us
the freedom to make certain kinds of changes locally, without causing ripples
through the entire system. We can replace how a module works inside, while
maintaining its contract. We can loosen its requirements, so it fits more use
cases. We can tighten its guarantees, so others can rely on it more strongly.

This approach scales well because modularity is *fractal*. At the top level of
a system, we have modules connected by interfaces. But if we zoom into any one
module, we can apply the same principles again: divide and conquer, define
clear boundaries, reason locally. The same thinking that helps us manage a
whole system can help us manage each of its parts.
