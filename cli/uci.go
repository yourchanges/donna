// Copyright (c) 2013-2014 by Michael Dvorkin. All Rights Reserved.
// Use of this source code is governed by a MIT-style license that can
// be found in the LICENSE file.

package cli

import(
	`github.com/michaeldv/donna`
	`bufio`
	`fmt`
	`os`
	`strings`
)

// Brain-damaged universal chess interface (UCI) protocol as described at
// http://wbec-ridderkerk.nl/html/UCIProtocol.html
func Uci() {
	var game *donna.Game
	var position *donna.Position

	// "uci" command handler.
	doUci := func(args []string) {
		fmt.Println(`Donna v1.0.0 Copyright (c) 2014 by Michael Dvorkin. All Rights Reserved.`)
		fmt.Println(`id author Michael Dvorkin`)
		fmt.Println(`uciok`)
	}

	// "ucinewgame" command handler.
	doUciNewGame := func(args []string) {
		game, position = nil, nil
	}

	// "isready" command handler.
	doIsReady := func(args []string) {
		fmt.Println(`readyok`)
	}

	// "position [startpos | fen ] [ moves ... ]" command handler.
	doPosition := func(args []string) {
		fmt.Printf("%q\n", args)

		// Make sure we've started the game since "ucinewgame" is optional.
		if game == nil || position == nil {
			game = donna.NewGame().Set(`cache`, 64, `movetime`, 5000) // 5s per move.
			position = game.Start()
		}

		switch args[0] {
		case `startpos`:
			args = args[1:]
			position = donna.NewInitialPosition(game)
		case `fen`:
			fen := []string{}
			for _, token := range args[1:] {
				args = args[1:] // Shift the token.
				if token == `moves` {
					break
				}
				fen = append(fen, token)
			}
			fmt.Printf("fen: %s\n", strings.Join(fen, ` `))
			position = donna.NewPositionFromFEN(game, strings.Join(fen, ` `))
		default: return
		}

		fmt.Printf("args: %q\n%s\n", args, position)
		if position != nil && len(args) > 0 && args[0] == `moves` {
			for _, move := range args[1:] {
				args = args[1:] // Shift the move.
				position = position.MakeMove(donna.NewMoveFromNotation(position, move))
			}
		}
		fmt.Printf("%s\n", position)
	}

	// "go [[wtime winc | btime binc ] movestogo] | depth | movetime"
	doGo := func(args []string) {
		fmt.Printf("%q\n", args)
	}

	// Stop calculating as soon as possible.
	doStop := func(args []string) {
	}

	// Quit the program as soon as possible.
	doQuit := func(args []string) {
		fmt.Printf("%q\n", args)
	}
	var commands = map[string]func([]string){
		`isready`: doIsReady,      
		`uci`: doUci,
		`ucinewgame`: doUciNewGame,
		`position`: doPosition,
		`go`: doGo,
		`stop`: doStop,
		`quict`: doQuit,
	}

 	bio := bufio.NewReader(os.Stdin)
	for {
		command, _ := bio.ReadString('\n')
		args := strings.Split(command[:len(command)-1], ` `)
		if args[0] == `quit` {
			break
		}
		if handler, found := commands[args[0]]; found {
			handler(args[1:])
		}
	}
}
