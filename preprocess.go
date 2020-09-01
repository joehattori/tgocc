package main

var macros = make(map[string][]token)

func (p *parser) preprocess() {
	var output []token
	for !p.isEOF() {
		if p.consume("\n") {
			continue
		}
		cur := p.toks[0]
		if id, ok := p.consumeID(); ok {
			if macro, ok := macros[id]; ok {
				output = append(output, macro...)
			} else {
				output = append(output, cur)
			}
			continue
		}

		if !p.consume("#") {
			output = append(output, cur)
			p.popToks()
			continue
		}

		if p.consume("define") {
			var def []token
			id := p.expectID()
			for !p.isEOF() {
				if p.consume("\n") {
					break
				}
				cur := p.toks[0]
				p.popToks()
				def = append(def, cur)
			}
			macros[id] = def
		} else if p.consume("include") {
		}
	}
	p.toks = append(output, p.toks[0])
}
