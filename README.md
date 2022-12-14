# Probe Memory Model

This is a tool which implements a few of the memory model probes discussed in [Russ Cox's fantastic "Hardware Memory Models" blog post](https://research.swtch.com/hwmm).

This tool is technically probing the combination of the hardware _and_ compiler which weakens the conclusion quite a bit. For example, it's not clear if a result is really due to the hardware or a compiler optimization.

That being said, it still achieves its goal of illustrating examples that might surprise a programmer and help to refine their idea that multithreaded code simply means "interleaving each threads' code".

I personally find one practical takeaway from all this to be:

In a large, multithreaded program you really don't know the kind of subtle behavior which can occur when not properly using synchronization primitives.

It also means that programming language designers are doing _incredible_ work abstracting these subtle behaviors away from us.

## Usage

Because this is architecture (_and possibly compiler version specific_), you'll want to build and run this on your own.

## Probes

_Note: Any "no" result below just means that I have not detected the tested behavior and should not be taken as proof that it cannot happen!_

### mp - Message passing

Probes whether memory writes are exchanged "at will".

```txt
Proc 1        Proc 2
x = 1         r1 = y
y = 1         r2 = x

Can we see?
r1 = 1
r2 = 0
```

My personal experience has been:
* Intel MBP + Go 1.18.1: No
* M1 MBP + Go 1.18.1: Yes!

### bw - Buffered writes

Probes whether memory writes are buffered in a write queue.

```txt
Proc 1        Proc 2
 x = 1         y = 1
r1 = y        r2 = x

Can we see?
r1 = 0
r2 = 0
```

My personal experience has been:
* Intel MBP + Go 1.18.1: Yes!
* M1 MBP + Go 1.18.1: Yes!

### iriw - Independent reads of independent writes

Probes whether procs can have distinct read orders of independent writes.

```txt
Proc 1        Proc 2        Proc 3        Proc 3
x = 1         y = 1         r1 = x        r3 = y
                            r2 = y        r4 = x

Can we see?
r1 = 1
r2 = 0
r3 = 1
r4 = 0
```

My personal experience has been:
* Intel MBP + Go 1.18.1: No
* M1 MBP + Go 1.18.1: No

### n6 - x86 violation of TLO+CC memory model (Paul Loewenstein)

Probe by Paul Loewenstein to show x86 violates TLO+CC memory model.

```txt
Proc 1        Proc 2
 x = 1        y = 1
r1 = x        x = 2
r2 = y

Can we see?
r1 = 1
r2 = 0
 x = 1
```

My personal experience has been:
* Intel MBP + Go 1.18.1: Yes!
* M1 MBP + Go 1.18.1: Yes!

Just to clarify why this test is interesting:

Observing `r1 = 1, r2 = 0, x = 1` provides evidence against a "simple interleaved instruction" model for the following reason:

Suppose our system does obey a "simple interleaved instruction" model.

Since `x = 1`, we know Proc 2 must have finished before Proc 1 since its last instruction sets `x = 2`.

That implies it already set `y = 1` hence `r2 = y = 1` which is a contradiction.

Hence, our system could not have followed a "simple interleaved instruction" model.

### rb - Read buffering

Probes whether reads can be buffered until after another procs writes.

```txt
Proc 1        Proc 2
r1 = x        r2 = y
 y = 1         x = 1

Can we see?
r1 = 1
r2 = 1
```

My personal experience has been:
* Intel MBP + Go 1.18.1: No
* M1 MBP + Go 1.18.1: No
