package truco

import (
	"errors"
	"fmt"
)

const MaxPlayers = 4
const NumCardsInHand = 3

// This file contains all the actions a user can perform on a game. It is the only file that can manipulate the global
// state

func CreateGameAndAddToGames(name string) (*Game, error) {
	newGame, err := createGame(Games, name)

	if err != nil {
		return nil, err
	}

	Games = append(Games, *newGame)

	notifyGameListChangeListeners()

	return newGame, nil
}

func DeleteGame(gameId string) error {
	index, _, err := FindGameWithId(gameId)

	if err != nil {
		return err
	}

	Games = append(Games[:index], Games[index+1:]...)

	notifyGameListChangeListeners()
	notifyGameChangeListeners(gameId)

	return nil
}

func FindGameWithId(id string) (int, *Game, error) {
	gameIndex := -1
	for index, game := range Games {
		if game.Id == id {
			gameIndex = index
		}
	}

	if gameIndex == -1 {
		return 0, nil, errors.New("game with id \"" + id + "\" not found")
	}

	return gameIndex, &Games[gameIndex], nil
}

func FindPlayerWithId(game Game, id string) (int, *Player, error) {
	playerIndex := -1
	for index, player := range game.Players {
		if player.Id == id {
			playerIndex = index
		}
	}

	if playerIndex == -1 {
		return 0, nil, errors.New("player with id \"" + id + "\" not found in game " + game.Id)
	}

	return playerIndex, &game.Players[playerIndex], nil
}

func CreatePlayer(gameId string, name string) (*Player, error) {
	gameIndex, game, err := FindGameWithId(gameId)

	if err != nil {
		return nil, err
	}

	if len(game.deck) < NumCardsInHand {
		return nil, errors.New("deck not big enough to make a new hand")
	}

	if len(game.Players) == MaxPlayers {
		return nil, errors.New("no more new players can be added")
	}

	newPlayer, err := createPlayer(name)

	if err != nil {
		return nil, err
	}

	game.Players = append(game.Players, *newPlayer)

	Games[gameIndex] = *game

	notifyPlayerListChangeListeners(gameId)
	notifyGameChangeListeners(gameId)

	return newPlayer, nil
}

func DeletePlayer(gameId string, playerId string) error {
	gameIndex, _, err := FindGameWithId(gameId)

	if err != nil {
		return err
	}

	playerIndex, _, err := FindPlayerWithId(Games[gameIndex], playerId)

	if err != nil {
		return err
	}

	players := Games[gameIndex].Players

	Games[gameIndex].Players = append(players[:playerIndex], players[playerIndex+1:]...)

	notifyPlayerListChangeListeners(gameId)
	notifyGameChangeListeners(gameId)

	return nil
}

func PlayCard(gameId string, playerId string, card Card) error {
	gameIndex, game, err := FindGameWithId(gameId)

	if err != nil {
		return err
	}

	playerIndex, player, err := FindPlayerWithId(*game, playerId)

	if err != nil {
		return err
	}

	index, err := findCardIndex(player.Hand, card)

	if err != nil {
		return err
	}

	for _, value := range game.Players[playerIndex].CardIndicesPlayed {
		if value == index {
			return errors.New(
				fmt.Sprintf("card %+v already played in this hand, cannot play card again", card),
			)
		}
	}

	cardIndicesPlayed := game.Players[playerIndex].CardIndicesPlayed
	game.Players[playerIndex] = Player{
		Id:                game.Players[playerIndex].Id,
		Name:              game.Players[playerIndex].Name,
		Hand:              player.Hand, // Updated value
		CardIndicesPlayed: append(cardIndicesPlayed, index),
	}

	newGame := Game{
		Name:    game.Name,
		Id:      gameId,
		Players: game.Players,
		deck:    game.deck,
	}

	Games[gameIndex] = newGame

	notifyGameChangeListeners(gameId)

	return nil
}

func DealCards(gameId string) error {
	gameIndex, game, err := FindGameWithId(gameId)

	if err != nil {
		return err
	}

	if len(game.deck) < NumCardsInHand*len(game.Players) {
		return errors.New("deck not big enough to deal cards")
	}

	newDeck := game.deck
	for playerId, player := range game.Players {
		player, deck := dealPlayerIn(newDeck, &player)
		newDeck = deck

		game.Players[playerId] = *player
	}

	updatedGame := Game{
		Name:    game.Name,
		Id:      game.Id,
		Players: game.Players,
		deck:    newDeck,
	}

	Games[gameIndex] = updatedGame

	notifyGameChangeListeners(gameId)

	return nil
}

