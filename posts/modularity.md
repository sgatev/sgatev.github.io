---
title: Modularity
date: 2025-08-01
---

Some systems are just too big for anyone to fully understand. So how do
we still build and work on them? The time-tested strategy is divide and
conquer: split the system into smaller pieces we can manage, then put
them back together to solve the bigger problem. But splitting alone
isn't enough. The value comes from *how* we split it and *how* we piece
it back together.

Good modular design is all about *local reasoning*. That means we should
be able to fully understand what a module of a system does without
having to figure out the whole thing. We should be able to tweak one
piece without forcing changes everywhere else. For this to work, we need
a clear, stable boundary between modules — a well-defined *interface*.
Interfaces bring everything together. More than that: they let us make a
few specific assumptions and ignore the details of the rest of the
system.

An interface can be viewed as a *contract* between suppliers and
consumers. It outlines what suppliers need from consumers and what they
promise to give in return. It's important that we take extra care to
specify the contract of each interface well, and to make sure that all
modules in the system respect these contracts. That will enable us to
build strong and scalable systems.

Modular systems are *fractal*. At a high level they are  modules
connected by interfaces. But if we zoom in on a specific module, we can
apply the same strategy: divide and conquer!
