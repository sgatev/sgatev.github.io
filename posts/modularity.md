---
title: Modularity
date: 2025-08-01
---

Some systems are just too big for anyone to fully understand. So how do
we still build and work with them? The time-tested strategy is divide
and conquer: split the system into smaller pieces we can manage, then
put them back together to solve the bigger problem. But splitting alone
isn't enough. The value comes from *how* we split it and *how* we piece
it back together.

Good modular design is all about *local reasoning*. That means we should
be able to fully understand what a module of a system does without
having to figure out the whole thing. We should be able to tweak one
piece without forcing changes everywhere else. For this to work well,
we need a clear, stable boundary between modules—a well-defined
*interface*. Interfaces bring everything together, letting us make a few
specific assumptions and ignore the details of the rest of the system.

An interface can be viewed as a *contract* between suppliers and
consumers. It outlines what suppliers need from consumers and what they
promise to give in return. It's important that we take extra care to
specify the contract of each interface well, and to make sure that all
modules in the system respect these contracts. We can’t rely on unit
tests or system tests alone to enforce them.

In a modular system, contracts are what let us work with confidence.
They give us the freedom to replace how a module works inside, to
loosen its requirements so it fits more use-cases, or to tighten its
guarantees so others can rely on it more strongly. These kinds of
changes remain local and do not ripple through the entire system.

Modular systems are *fractal*. At a high level they consist of modules
that are connected by interfaces. If we zoom in and look at a specific
module, we see that we can apply the same strategy to simplify it
further. Divide and conquer!
