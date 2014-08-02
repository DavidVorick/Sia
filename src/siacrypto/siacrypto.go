// Sia uses a specific set of cryptographic functions across the entire project.
// All of those functions are defined and implemented within the siacrypto
// package to maintain consistency.
//
// Currently, siacrypto is in a state of 'hotfix', meaning we intend to swap
// out the existing functions for NaCl or libsodium.
package siacrypto
