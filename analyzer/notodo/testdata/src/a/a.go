package a

// TODO fix later // want `TODO marker is forbidden`
func Bad1() {}

// FIXME this is bad // want `FIXME marker is forbidden`
func Bad2() {}

// HACK relying on env // want `HACK marker is forbidden`
func Bad3() {}

// TODO(@alice): fix later // want `TODO marker is forbidden`
func Now1() {}

// FIXME PROJ-123 implement // want `FIXME marker is forbidden`
func Now2() {}

// see the TODO list in the wiki for the full backlog
func Fine() {}
