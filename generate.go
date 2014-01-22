// Copyright (c) 2013 by Michael Dvorkin. All Rights Reserved.
// Use of this source code is governed by a MIT-style license that can
// be found in the LICENSE file.

package donna

import() //(`sort`)

const (
        stepPrincipal = iota
        stepCaptures
        stepPromotions
        stepKillers
        stepRemaining
)

type MoveEx struct {
        move   Move
        score  int
}

type MoveList struct {
        position  *Position
        moves     [256]MoveEx
        ply       int
        head      int
        tail      int
        step      int
}

var moveList [MaxPly]MoveList

func (p *Position) StartMoveGen(ply int) (ml *MoveList) {
        ml = &moveList[ply]
        ml.position = p
        ml.moves = [256]MoveEx{}
        ml.ply = ply
        ml.head, ml.tail = 0, 0
        return
}

func (ml *MoveList) NextMove() (move Move) {
        if ml.head == ml.tail {
                return 0
        }
        move = ml.moves[ml.head].move
        ml.head++
        return
}

func (ml *MoveList) GenerateMoves() *MoveList {
        color := ml.position.color
        ml.pawnMoves(color)
        ml.kbrqMoves(color)
        ml.kingMoves(color)
        return ml
}

func (ml *MoveList) pawnMoves(color int) *MoveList {
        pawns := ml.position.outposts[Pawn(color)]

        for pawns != 0 {
                square := pawns.pop()
                targets := ml.position.targets[square]
                for targets != 0 {
                        target := targets.pop()
                        if target > H1 && target < A8 {
                                ml.moves[ml.tail].move = ml.position.pawnMove(square, target)
                                ml.tail++
                        } else { // Promotion.
                                m1, m2, m3, m4 := ml.position.pawnPromotion(square, target)
                                ml.moves[ml.tail].move = m1
                                ml.tail++
                                ml.moves[ml.tail].move = m2
                                ml.tail++
                                ml.moves[ml.tail].move = m3
                                ml.tail++
                                ml.moves[ml.tail].move = m4
                                ml.tail++
                        }
                }
        }
        return ml
}

func (ml *MoveList) kbrqMoves(color int) *MoveList {
	for _, kind := range [4]int{ KNIGHT, BISHOP, ROOK, QUEEN } {
	        outposts := ml.position.outposts[Piece(kind|color)]
	        for outposts != 0 {
	                square := outposts.pop()
	                targets := ml.position.targets[square]
	                for targets != 0 {
	                        target := targets.pop()
	                        ml.moves[ml.tail].move = NewMove(ml.position, square, target)
	                        ml.tail++
	                }
	        }
	}
        return ml
}

func (ml *MoveList) kingMoves(color int) *MoveList {
        var move Move
        king := ml.position.outposts[King(color)]
        if king != 0 {
                square := king.pop()
                targets := ml.position.targets[square]
                for targets != 0 {
                        target := targets.pop()
                        if square == homeKing[color] && Abs(square - target) == 2 {
                                move = NewCastle(ml.position, square, target)
                        } else {
                                move = NewMove(ml.position, square, target)
                        }
                        ml.moves[ml.tail].move = move
                        ml.tail++
                }
        }
        return ml
}


func (ml *MoveList) GenerateCaptures() *MoveList {
        for square, piece := range ml.position.pieces {
                if piece != 0 && piece.color() == ml.position.color {
                        ml.possibleCaptures(square, piece)
                }
        }
        return ml
}

func (ml *MoveList) possibleCaptures(square int, piece Piece) *MoveList {
        targets := ml.position.targets[square]

        for targets != 0 {
                target := targets.pop()
                capture := ml.position.pieces[target]
                if capture != 0 {
                        if !ml.position.isPawnPromotion(piece, target) {
                                ml.moves[ml.tail].move = NewMove(ml.position, square, target)
                                ml.tail++
                        } else {
                                for _,name := range([]int{ QUEEN, ROOK, BISHOP, KNIGHT }) {
                                        ml.moves[ml.tail].move = NewMove(ml.position, square, target).promote(name)
                                        ml.tail++
                                }
                        }
                } else if ml.position.flags.enpassant != 0 && target == ml.position.flags.enpassant {
                        ml.moves[ml.tail].move = NewMove(ml.position, square, target)
                        ml.tail++
                }
        }
        return ml
}

// All moves.
func (p *Position) Moves(ply int) (moves []Move) {
        for square, piece := range p.pieces {
                if piece != 0 && piece.color() == p.color {
                        moves = append(moves, p.possibleMoves(square, piece)...)
                }
        }
        moves = p.reorderMoves(moves, p.game.bestLine[0][ply], p.game.killers[ply])
        Log("%d candidates for %s: %v\n", len(moves), C(p.color), moves)
        return
}

func (p *Position) Captures(ply int) (moves []Move) {
        for i, piece := range p.pieces {
                if piece != 0 && piece.color() == p.color {
                        moves = append(moves, p.possibleCaptures(i, piece)...)
                }
        }
        if bestMove := p.game.bestLine[0][ply]; bestMove != 0 && bestMove.capture() != 0 {
                moves = p.reorderCaptures(moves, bestMove)
        } else {
                //sort.Sort(byScore{moves})
        }

        Log("%d capture candidates for %s: %v\n", len(moves), C(p.color), moves)
        return
}

// All moves for the piece in certain square. This might include illegal
// moves that cause check to the king.
func (p *Position) possibleMoves(square int, piece Piece) (moves []Move) {
        targets := p.targets[square]

        for targets != 0 {
                target := targets.pop()
                //
                // For regular moves each target square represents one possible
                // move. For pawn promotion, however, we have to generate four
                // possible moves, one for each promoted piece.
                //
                if !p.isPawnPromotion(piece, target) {
                        moves = append(moves, NewMove(p, square, target))
                } else {
                        for _,name := range([]int{ QUEEN, ROOK, BISHOP, KNIGHT }) {
                                candidate := NewMove(p, square, target).promote(name)
                                moves = append(moves, candidate)
                        }
                }
        }
        return
}

// All capture moves for the piece in certain square. This might include
// illegal moves that cause check to the king.
func (p *Position) possibleCaptures(square int, piece Piece) (moves []Move) {
        targets := p.targets[square]

        for targets != 0 {
                target := targets.pop()
                capture := p.pieces[target]
                if capture != 0 {
                        if !p.isPawnPromotion(piece, target) {
                                moves = append(moves, NewMove(p, square, target))
                        } else {
                                for _,name := range([]int{ QUEEN, ROOK, BISHOP, KNIGHT }) {
                                        candidate := NewMove(p, square, target).promote(name)
                                        moves = append(moves, candidate)
                                }
                        }
                } else if p.flags.enpassant != 0 && target == p.flags.enpassant {
                        moves = append(moves, NewMove(p, square, target))
                }
        }
        return
}

func (p *Position) reorderMoves(moves []Move, bestMove Move, goodMove [2]Move) []Move {
        var principal, killers, captures, promotions, remaining []Move

        for _, move := range moves {
                if len(principal) == 0 && bestMove != 0 && move == bestMove {
                        principal = append(principal, move)
                } else if move.capture() != 0 {
                        captures = append(captures, move)
                } else if move.promo() != 0 {
                        promotions = append(promotions, move)
                } else if (goodMove[0] != 0 && move == goodMove[0]) || (goodMove[1] != 0 && move == goodMove[1]) {
                        killers = append(killers, move)
                } else {
                        remaining = append(remaining, move)
                }
        }
        if len(killers) > 1 && killers[0] == goodMove[1] {
                killers[0], killers[1] = killers[1], killers[0]
        }

        //sort.Sort(byScore{captures})
        //sort.Sort(byScore{remaining})
        return append(append(append(append(append(principal, captures...), promotions...), killers...), remaining...))
}

func (p *Position) reorderCaptures(moves []Move, bestMove Move) []Move {
        var principal, remaining []Move

        for _, move := range moves {
                if len(principal) == 0 && move == bestMove {
                        principal = append(principal, move)
                } else {
                        remaining = append(remaining, move)
                }
        }
        //sort.Sort(byScore{remaining})
        return append(principal, remaining...)
}

// Sorting moves by their relative score based on piece/square for regular moves
// or least valuaeable attacker/most valueable victim for captures.
// type byScore struct {
//         moves []Move
// }
// func (her byScore) Len() int           { return len(her.moves)}
// func (her byScore) Swap(i, j int)      { her.moves[i], her.moves[j] = her.moves[j], her.moves[i] }
// func (her byScore) Less(i, j int) bool { return her.moves[i].score > her.moves[j].score }

func (p *Position) pawnMove(square, target int) Move {
        color := p.color

        if RelRow(square, color) == 1 && RelRow(target, color) == 3 {
                if p.isEnpassant(target, color) {
                        return NewEnpassant(p, square, target)
                } else {
                        return NewPawnJump(p, square, target)
                }
        }

        return NewMove(p, square, target)
}

func (p *Position) pawnPromotion(square, target int) (m1, m2, m3, m4 Move) {
        m1 = NewMove(p, square, target).promote(QUEEN)
        m2 = NewMove(p, square, target).promote(ROOK)
        m3 = NewMove(p, square, target).promote(BISHOP)
        m4 = NewMove(p, square, target).promote(KNIGHT)
        return
}

func (p *Position) isEnpassant(target, color int) bool {
        pawns := p.outposts[Pawn(color^1)] // Opposite color pawns.
        switch col := Col(target); col {
        case 0:
                return pawns.isSet(target + 1)
        case 7:
                return pawns.isSet(target - 1)
        default:
                return pawns.isSet(target + 1) || pawns.isSet(target - 1)
        }
        return false
}
