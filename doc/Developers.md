This document is for people who would like to contribute to the Sia project.

Security, readability, and maintainability are our primary goals. With the temporary exception of the lead developers, all code need to be reviewed for correctness, best practice, and usefulness.

All code needs to be audited for bugs and backdoors. It is not simply good enough, however, for someone to look at the code and declare that they cannot find bugs. They must also declare that they do not think bugs are likely to be contained within the code. That means the code needs to follow a long set of rules that keep things simple, low risk, and maintainable:

1. Pointers are only to be used if the intent is for the calling function to edit the existing object
2. All algorithms need documentation
3. All structs need documentation
4. Code must be simple to understand. Pull requests can be rejected because the code is simply difficult to read.

+ All functions should have an explanatory comment above them. For many functions, this comment will need to be lengthy. For many other functions, this comment need only have a few words. In all cases, the name of the function needs to appear in the comment.
+ All error messages should explain which function they are from. For example: "ReedSolomonEncode: received nil input!"
+ All functions must have corresponding tests. For now, these tests need only test the general use case of the function. Once Sia is in beta, these tests much check all corner cases, and must make a thorough attempt at breaking the function.
+ For all data structures, types of explicit sizes must be used. This means using 'int32' or 'int64' instead of 'int'.
+ Comments should have complete punctuation.
+ When saying the name of a variable within a comment, the variable should be in single quotes, for example 'exampleVariable'.
+ Be very careful with pointers.

Hooks:

GobHooks:
	Must always return error if there is nil input. Nil input should always be tested for.

Encoding is for data that stays on the computer, whether it is going to disk, or being hashed, or whatever. Encoding stays on the disk, has an arbitrary implementation, and is tracked by the protocol.
Marshalling is for objects that go over the wire, or have their hashes compared. Marshalling must follow an exact specification set out by the protocol.

Each time the weight of the quorum is changed in an upward direction, must check that AtomsPerQuorum is not being exceeded.

Objects need to be hashed before they are signed or verified. To prevent misuse and mistakes, the functions SignObject and VerifyObject should be used.

This is a work in progress.
