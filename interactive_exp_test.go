package cobra

// exports to private readerFactory var to mock the interactive tests
// ... to enable testing with mocked input
var ReaderFactory = &readerFactory

// exports to private selectionFactory var to mock the interactive tests
// ... to enable testing with mocked input
var SelectionFactory = &selectionFactory
