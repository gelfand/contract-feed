package core

import (
	"context"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gelfand/contract-feed/telegram"
)

type Coordinator struct {
	tg     *telegram.Client
	client *ethclient.Client
	signer types.Signer

	// headersCh is Ethereum headers channel.
	headersCh chan *types.Header
	txsCh     chan types.Transactions
	errCh     chan error
}

type Config struct {
	TelegramToken  string
	TelegramChatID int64
	RpcAddress     string
}

func NewCoordinator(ctx context.Context, cfg Config) (*Coordinator, error) {
	client, err := ethclient.DialContext(ctx, cfg.RpcAddress)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve connection with Ethereum RPC: %w", err)
	}
	chainID, err := client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve chainID: %v", err)
	}
	signer := types.LatestSignerForChainID(chainID)

	log.Printf("INFO: Successfully retrieved chain ID: %v", chainID)

	tg, err := telegram.NewClient(cfg.TelegramToken, cfg.TelegramChatID)
	if err != nil {
		return nil, err
	}

	return &Coordinator{
		tg:        tg,
		client:    client,
		signer:    signer,
		headersCh: make(chan *types.Header),
		txsCh:     make(chan types.Transactions),
		errCh:     make(chan error, 1),
	}, nil
}

func (c *Coordinator) startFilterer(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case txs := <-c.txsCh:
			for i := range txs {
				go func(tx *types.Transaction) {
					if tx.To() != nil {
						return
					}

					from, err := types.Sender(c.signer, tx)
					if err != nil {
						// it's unexpected error, so in those cases we better fully exit.
						c.errCh <- fmt.Errorf("could not recovery transaction sender: %w", err)
						return
					}

					addr := crypto.CreateAddress(from, tx.Nonce())
					if !c.isToken(ctx, addr) {
						return
					}

					fmt.Printf("Token: %v\n", addr)
					token, err := c.getTokenData(addr)
					if err != nil {
						log.Printf("DBUG: %v", err)
						// just ignore this.
						return
					}
					fmt.Println(token.Address, token.Symbol)

					// Telegram can't send non-unicode messages, but we might have cases where
					// token name and/or token symbol is not part of unicode, so we need to handle it.
					if err = c.tg.SendMsg(token.ToMsg()); err != nil {
						log.Printf("Could not send token data: %v", err)
					}
				}(txs[i])
			}
		}
	}
}

func (c *Coordinator) Run(ctx context.Context) error {
	go c.startFilterer(ctx)

	sub, err := c.client.SubscribeNewHead(ctx, c.headersCh)
	if err != nil {
		return fmt.Errorf("could not subscribe to new Ethereum headers: %w", err)
	}
	defer sub.Unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-c.errCh:
			return err
		case err := <-sub.Err():
			return fmt.Errorf("could not continue handling of headers subscription: %w", err)
		case header := <-c.headersCh:
			block, err := c.client.BlockByHash(ctx, header.Hash())
			if err != nil {
				log.Printf("TRACE: could not retrieve block by hash: %v", err)
				continue
			}
			fmt.Println(block.Number(), len(block.Transactions()))
			c.txsCh <- block.Transactions()
		}
	}
}
