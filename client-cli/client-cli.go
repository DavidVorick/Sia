// Cli Conventions: 'help' is always at the top, followed by quit. All other
// actions are listed in alphabetic order after these two.
//
// The client cli is modal, currently having two states. The first state is the
// 'home' state, where you can request wallets, load wallets, and do general
// high level management of your wallets. The second state is a wallet state,
// which performs actions against a specific wallet. The third state is the
// server state, where you can manage groups of participants that are providing
// storage to the network.
//
// Function conventions: helper/misc functions are listed alphabetically at the
// top, followed by all walkthroughs, followed by all mode swtiching functions,
// followed by displayXXXHelp(), followed by pollXXX. See home.go for an
// example.
package main
