package main

var style = `
html{
    background: rgb(248,241,228);
    font-family: Iowan, Palatino, Times, 'Times New Roman', serif!important;
    -webkit-user-select: text;
    -webkit-font-smoothing: subpixel-antialiased;
	text-rendering: optimizeLegibility;
	pointer-events: auto;
	line-height: 28px;
}

form.search{
	text-align: center;
}
form.search input {
	font-family: Iowan;
	width: 80%;
	line-height: 32px;
	padding: 8px;
	font-size: 18pt;
}

.scrollable {
	-webkit-font-smoothing: subpixel-antialiased;
	-webkit-user-select: text;
	max-width: 100%;
	overflow-x: scroll;
	pointer-events: auto;
	text-align: start;
	text-rendering: optimizelegibility;
}

.scrollable table {
	width: 100%;
}

.scrollable th, .scrollable tr {
	border: 0px;
}
.scrollable th {
	white-space:nowrap;
	line-height: 32px;
	background-color: rgba(154, 128, 92, 0.0588235);
}

.scrollable th, .scrollable td {
	border-bottom-color: rgb(230, 218, 202);
	border-bottom-style: solid;
	border-bottom-width: 1px;
	border-collapse: collapse;
}

#container {
	padding: 20px;
}

img {
	max-width:100%;
}

a{
    color: rgb(82,129,197)!important;
}

body {
    padding-bottom:1px;
    color: rgb(72,41,19)!important;
	font-size: 14pt;
    max-width: 800px;
    margin: 0 auto!important;
word-wrap: break-word;
}

#page {
}

#topbar {
    display: none;
}

h1 {
	font-size: 20pt;
}

h2 {
    background: transparent!important;;
    font-size:14pt!important;
}

pre span {
    font-family: "m+ 2m"!important;
}

pre {
    background: #f4f4f4!important;
    background: rgb(244,236,221)!important;
    font-family: "m+ 2m"!important;
	padding: 8px;
	border-radius: 4px;

	-webkit-font-smoothing: subpixel-antialiased;
	-webkit-user-select: text;
	max-width: 100%;
	overflow-x: scroll;
	pointer-events: auto;
}
`
