# Pokedex CLI

A command-line Pokedex built with Go — catch, inspect, and list Pokémon right from your terminal.  
Inspired by the classic Pokémon games: *Gotta catch 'em all!* 🐾

## Features
- **Catch Pokémon**: Attempt to catch wild Pokémon using `catch <pokemon>`.
- **Inspect Pokémon**: View details about Pokémon you’ve caught.
- **View Pokedex**: List all Pokémon you’ve captured so far.
- **Map Navigation**: Explore regions and discover new Pokémon.
- **Command History Navigation**: Quickly cycle through previously entered commands using the ↑ / ↓ arrow keys for faster input.

## Example Usage

```
Pokedex > catch pidgey
Throwing a Pokeball at pidgey...
pidgey was caught!
You may now inspect it with the inspect command.

Pokedex > catch caterpie
Throwing a Pokeball at caterpie...
caterpie was caught!
You may now inspect it with the inspect command.

Pokedex > pokedex
Your Pokedex:
 - pidgey
 - caterpie
```

## Installation

1. **Clone the repository**:
   ```
   git clone git@github.com:<your-username>/pokedex-cli.git
   cd pokedex-cli
   ```

2. **Build the CLI**:
    ```
    go build -o pokedex
    ```

3. **Run it**:
    ```
    ./pokedex
    ```


## Requirements
 * Go 1.20+
 * Internet connection (for fetching Pokémon data via API)