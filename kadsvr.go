package main

// KadSvr is the main entry point for the Kademlia server. It wraps
// the distributed hash table implementation with an http.Handler, which
// provides an interface for clients to execute read/write operations
type KadSvr struct {
	hash *Dht
}

/*
func (k *KadSvr) Set(ctx *fiber.Ctx) {
	value := ctx.Params("Value")
	//req := k.FormStoreMsg(value)
	//resp := new(StoreResp)
	//k.hash.Store(req, resp)

	//msg := fmt.Sprintf("Storing value with key: %v\n", string(resp.Key))
	ctx.SendString(msg)
}

func (k *KadSvr) Get(ctx *fiber.Ctx) {
	key := ctx.Params("Key")
	//req := k.FormFindKeyMsg(key)
	//resp := new(FindValueResult)
	//k.hash.FindValue(req, resp)

	msg := fmt.Sprintf("The corresponding value for the key %s is: %v", key, resp.Value)
	ctx.SendString(msg)
}

func (k *KadSvr) AddRoutes() {
	k.handler.Post("/Set/:Value", k.Set)

	k.handler.Get("/Get/:Key", k.Get)
}
*/
