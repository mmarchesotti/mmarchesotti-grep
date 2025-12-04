# Go Grep: A Custom `grep` Implementation

This project is a from-scratch implementation of the `grep` command-line utility in Go. It is designed to explore the principles of regular expression engines, including parsing, compilation, and execution using a Non-deterministic Finite Automaton (NFA).

## Origin & Acknowledgments

This project initially started as a solution to the [CodeCrafters "Build Your Own Grep"](https://codecrafters.io/challenges/grep) challenge. However, halfway through the process, the implementation branched out significantly from the guided path. I chose to develop a custom testing strategy and implement my own architectural ideas—specifically the hybrid NFA/Backtracking engine—to better understand the nuances of regex engine design beyond the scope of the original tutorial.

## Features

* **Pattern Matching**: Search for regex patterns in files or standard input.
* **File & Stdin Support**: Accepts a list of files to search or reads from `stdin` when no files are provided.
* **Recursive Search**: Use the `-r` flag to recursively search for patterns within a directory.
* **Hybrid Engine**:
  * **NFA Engine**: Uses Thompson's construction for O(n) performance on standard patterns.
  * **Backtracking Engine**: Automatically engages for patterns containing backreferences, combining NFA fragments with recursive checks to handle stateful matching.

## Supported Regex Syntax

The engine supports a solid subset of common ERE (Extended Regular Expression) features:

| Feature | Syntax | Example | Description |
| :--- | :--- | :--- | :--- |
| Literals | `a`, `b`, `1` | `cat` | Matches the exact character sequence. |
| Character Classes | `\d`, `\w` | `\d{3}` | Matches digits or word characters. |
| Character Sets | `[...]` | `[abc]` | Matches any character in the set. |
| Negated Sets | `[^...]` | `[^0-9]` | Matches any character not in the set. |
| Wildcard | `.` | `a.c` | Matches any character except newline. |
| Quantifiers | `*`, `+`, `?` | `a*`, `b+`, `c?` | Match zero-or-more, one-or-more, or zero-or-one times. |
| Alternation | `|` | `cat\|dog` | Matches either "cat" or "dog". |
| Grouping | `(...)` | `(ab)+` | Groups expressions for quantifiers or alternation. |
| Backreferences | `\1`, `\2`, ... | `(a)\1` | Matches the exact text captured by a previous group. |
| Positional Anchors | `^`, `$` | `^start`, `end$` | Matches the beginning or end of a line. |

## Architecture

This project is built using a multi-stage, compiler-inspired pipeline to process and execute regular expressions. This design is robust, modular, and easy to extend.

The flow is as follows:

1. **Lexer (`lexer.go`)**: The raw regex string is fed into the lexer, which breaks it down into a flat sequence of tokens (e.g., `LITERAL`, `KLEENE_CLOSURE`, `BACKREFERENCE`).

2. **Parser (`parser.go`)**: The stream of tokens is organized into a hierarchical **Abstract Syntax Tree (AST)**. The AST represents the grammatical structure and precedence of the regex operators.

3. **Hybrid Execution Strategy**:
   * **Standard Compilation**: For patterns without backreferences, the AST is compiled into a **Non-deterministic Finite Automaton (NFA)** using Thompson's construction (`build_nfa.go`). This ensures linear-time execution regardless of complexity.
   * **Backtracking Logic (`backtrack.go`)**: When backreferences are detected, the engine switches to a recursive strategy. It splits the pattern into NFA fragments (handled by the NFA Simulator) and verifies backreferences by comparing captured text against the input stream, backtracking if a match fails.

4. **NFA Simulator (`nfa_simulator.go`)**: The core execution unit that runs NFA fragments against the input text. It steps through the input character by character, tracking all possible active states.

This hybrid approach allows the engine to remain highly efficient for standard patterns while still supporting complex features like backreferences when necessary.

## Usage

### Building

To build the executable, run the following command from the project's root directory:

```sh
go build -o mygrep ./cmd/mygrep
````

### Examples

**Search for a pattern in files:**

```sh
./mygrep 'pattern' file1.txt file2.txt
```

**Search using backreferences:**

```sh
# Matches repeated words like "the the" or "is is"
./mygrep '(\w+) \1' file.txt
```

**Recursive search within a directory:**

```sh
./mygrep -r 'TODO' ./project_directory
```
