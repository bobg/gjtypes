# Gjtypes - Convert JSON data into the Go types needed to parse it

This is gjtypes,
a command that parses JSON data on standard input
and produces the Go types needed to parse that data
on standard output.

## Installation

```sh
go install github.com/bobg/gjtypes@latest
```

## Usage

```sh
gjtypes < INPUT
```

where INPUT is a file of JSON data.

The resulting Go code may include generated type names − S001, S002, and so on −
that you’ll probably want to rename to be meaningful.

## Example

Given the following JSON input:

```javascript
{"title": "Scott Pilgrim vs. The World",
 "director": "Edgar Wright",
 "year": 2010,
 "stars": ["Michael Cera",
           "Mary Elizabeth Winstead",
           "Ellen Wong",
           "Chris Evans",
           "Aubrey Plaza",
           "Anna Kendrick",
           "Brie Larson"],
 "songs": [{"title": "Teenage Dream",
            "artist": "T. Rex"},
           {"title": "Black Sheep",
            "artist": "Metric"}]}
```

gjson will emit:

```go
var data *S001 // Unmarshal into this type.

type S001 struct {
	Director string   `json:"director,omitempty"`
	Songs    []*S002  `json:"songs,omitempty"`
	Stars    []string `json:"stars,omitempty"`
	Title    string   `json:"title,omitempty"`
	Year     int64    `json:"year,omitempty"`
}

type S002 struct {
	Artist string `json:"artist,omitempty"`
	Title  string `json:"title,omitempty"`
}
```

In this case you’d probably want to rename `S001` to `Movie`
and `S002` to `Song`.
