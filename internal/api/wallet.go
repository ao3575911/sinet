package api

import (
	"database/sql"
	"net/url"
	"strings"
	"time"

	"sinet/internal/wallet"
)

func (c *call) wallets() (any, error) {
	items, err := c.many(`SELECT json_build_object('address',address,'chain_id',chain_id,'verified_at',verified_at,'status','ownership_verified') FROM wallet_links WHERE account_id=$1 ORDER BY verified_at,address`, c.user)
	if err != nil {
		return nil, err
	}
	result := items.(map[string]any)
	result["supported_chain_id"] = 1
	result["mode"] = "ownership_link"
	result["transactions_available"] = false
	result["balances_available"] = false
	return result, nil
}

func (c *call) walletChallenge() (any, error) {
	if err := c.activated(); err != nil {
		return nil, err
	}
	var in struct {
		Address string `json:"address"`
	}
	if err := decode(c.r, &in); err != nil {
		return nil, err
	}
	address, err := wallet.Address(in.Address)
	if err != nil {
		return nil, bad("invalid_ethereum_address")
	}
	if err := c.limit("wallet-challenge:"+c.user, 5, 15*time.Minute); err != nil {
		return nil, err
	}
	var count int
	if err := c.tx.QueryRowContext(c.r.Context(), `SELECT count(*) FROM wallet_links WHERE account_id=$1`, c.user).Scan(&count); err != nil {
		return nil, err
	}
	if count >= 5 {
		return nil, fail(429, "wallet_limit_5")
	}
	nonce := random(16)
	now := time.Now().UTC().Truncate(time.Second)
	expires := now.Add(5 * time.Minute)
	var handle string
	if err := c.tx.QueryRowContext(c.r.Context(), `SELECT handle FROM profiles WHERE account_id=$1`, c.user).Scan(&handle); err != nil {
		return nil, err
	}
	message, err := wallet.Message(c.s.PublicURL, address, handle, nonce, now, expires)
	if err != nil {
		return nil, fail(503, "wallet_origin_unavailable")
	}
	if err := c.exec(`INSERT INTO wallet_challenges(account_id,nonce,address,message,expires_at,origin) VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT(account_id) DO UPDATE SET nonce=excluded.nonce,address=excluded.address,message=excluded.message,expires_at=excluded.expires_at,origin=excluded.origin,attempts=0,consumed=false`, c.user, nonce, address, message, expires, strings.TrimRight(c.s.PublicURL, "/")); err != nil {
		return nil, err
	}
	fragment := url.Values{"address": {wallet.Checksum(address)}, "account": {handle}, "nonce": {nonce}, "message": {message}}
	return map[string]any{"nonce": nonce, "address": wallet.Checksum(address), "chain_id": 1, "account": handle, "message": message, "expires_at": expires, "purpose": "link_wallet_ownership", "signing_method": "personal_sign", "signing_url": strings.TrimRight(c.s.PublicURL, "/") + "/wallet/link#" + fragment.Encode()}, nil
}

func (c *call) verifyWallet() (any, error) {
	var in struct {
		Nonce     string `json:"nonce"`
		Signature string `json:"signature"`
	}
	if err := decode(c.r, &in); err != nil {
		return nil, err
	}
	if len(in.Nonce) != 32 || len(in.Signature) > 132 {
		return nil, bad("invalid_wallet_proof")
	}
	var address, message string
	err := c.tx.QueryRowContext(c.r.Context(), `SELECT address,message FROM wallet_challenges WHERE account_id=$1 AND nonce=$2 AND origin=$3 AND expires_at>clock_timestamp() AND NOT consumed AND attempts<5 FOR UPDATE`, c.user, in.Nonce, strings.TrimRight(c.s.PublicURL, "/")).Scan(&address, &message)
	if err == sql.ErrNoRows {
		return nil, fail(403, "wallet_challenge_unavailable")
	}
	if err != nil {
		return nil, err
	}
	if err := c.exec(`UPDATE wallet_challenges SET attempts=attempts+1,consumed=(attempts+1>=5) WHERE account_id=$1`, c.user); err != nil {
		return nil, err
	}
	if err := wallet.Verify(address, message, strings.TrimSpace(in.Signature)); err != nil {
		return nil, fail(403, "wallet_proof_invalid")
	}
	if err := c.exec(`UPDATE wallet_challenges SET consumed=true WHERE account_id=$1`, c.user); err != nil {
		return nil, err
	}
	var count int
	if err := c.tx.QueryRowContext(c.r.Context(), `SELECT count(*) FROM wallet_links WHERE account_id=$1`, c.user).Scan(&count); err != nil {
		return nil, err
	}
	if count >= 5 {
		return nil, fail(429, "wallet_limit_5")
	}
	var linkedAccount string
	err = c.tx.QueryRowContext(c.r.Context(), `SELECT account_id FROM wallet_links WHERE chain_id=1 AND address=$1`, address).Scan(&linkedAccount)
	if err == nil {
		return nil, fail(409, "wallet_already_linked")
	}
	if err != sql.ErrNoRows {
		return nil, err
	}
	if err := c.exec(`INSERT INTO wallet_links(account_id,chain_id,address,proof_hash) VALUES($1,1,$2,$3)`, c.user, address, hash(message)); err != nil {
		return nil, err
	}
	if err := c.exec(`INSERT INTO identity_audit(account_id,action) VALUES($1,'wallet_link')`, c.user); err != nil {
		return nil, err
	}
	return map[string]any{"address": wallet.Checksum(address), "chain_id": 1, "status": "ownership_verified"}, nil
}

func (c *call) unlinkWallet() (any, error) {
	var in struct {
		Confirm bool `json:"confirm"`
	}
	if err := decode(c.r, &in); err != nil {
		return nil, err
	}
	if !in.Confirm {
		return nil, bad("unlink_confirmation_required")
	}
	address, err := wallet.Address(c.r.PathValue("address"))
	if err != nil {
		return nil, bad("invalid_ethereum_address")
	}
	var removed string
	if err := c.tx.QueryRowContext(c.r.Context(), `DELETE FROM wallet_links WHERE account_id=$1 AND chain_id=1 AND address=$2 RETURNING address`, c.user, address).Scan(&removed); err != nil {
		return nil, err
	}
	if err := c.exec(`UPDATE wallet_challenges SET consumed=true WHERE account_id=$1 AND address=$2`, c.user, address); err != nil {
		return nil, err
	}
	if err := c.exec(`INSERT INTO identity_audit(account_id,action) VALUES($1,'wallet_unlink')`, c.user); err != nil {
		return nil, err
	}
	return map[string]any{"unlinked": wallet.Checksum(removed), "funds_moved": false}, nil
}
