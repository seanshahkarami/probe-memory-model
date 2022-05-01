# Probe Memory Model

This is a tool which implements a few of the memory model probes discussed in Russ Cox's "Hardware Memory Models" blog post.

Technically, I'm probing the combination of hardware + compiler, but this still illustrates that the behavior of a program isn't simply "interleaving the code" between threads in some random order.

## Probes

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
* Intel MBP: No
* M1 MBP: Yes!

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
* Intel MBP: Yes!
* M1 MBP: Yes!

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
* Intel MBP: No
* M1 MBP: TBD

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
* Intel MBP: Yes!
* M1 MBP: TBD

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
* Intel MBP: No
* M1 MBP: TBD
