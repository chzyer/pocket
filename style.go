package main

var style = `
html{
    background: rgb(248,241,228)!important;
    font-family: Iowan, Palatino, Times, 'Times New Roman', serif!important;
    -webkit-user-select: text;
    -webkit-font-smoothing: subpixel-antialiased;
	text-rendering: optimizeLegibility;
	pointer-events: auto;
	line-height: 28px;
}

html, button {
    font-family: Iowan, Palatino, Times, 'Times New Roman', serif!important;
}

blockquote {
	margin-left: 8px;
	padding-left: 28px;
	border-left: 4px solid rgb(102, 71, 49);
}

button {
	padding: 8px;
	background: rgb(244, 236, 221);
	color: rgb(72, 41, 19);
	font-size: 14pt;
	border: 0px;
	cursor: pointer;
}
button:hover {
	background: rgb(234, 226, 211);
}

hr {
	border: none;
	border-bottom: 1px solid rgb(102, 71, 49);
}

h1 {
	font-size: 20pt;
}

h1#title {
	border: none;
}

h2 {
    background: transparent!important;
    font-size:16pt!important;
}

h3 {
    background: transparent!important;
    font-size:14pt!important;
}

h1, h2, h3 {
	border-bottom: 1px solid rgb(102, 71, 49);
}

form.search{
	text-align: center;
}
form.search input {
	font-family: Iowan;
	width: 90%;
	outline: none;
	border: 3px solid rgb(172,141,119);
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
	background: transparent!important;
    color: rgb(72,41,19)!important;
	font-size: 14pt;
    max-width: 800px;
    margin: 0 auto!important;
	word-wrap: break-word;
}

.alignright {
	float: right;
	text-align: right;
	margin-left: 10px;
}

div.rich_media_inner {
	background: transparent!important;
}

#page {
}

#topbar {
    display: none;
}


pre, pre span, pre code {
    font-family: "m+ 2m", "monaco"!important;
}

pre {
    background: #f4f4f4!important;
    background: rgb(244,236,221)!important;
	padding: 8px;
	border-radius: 4px;

	-webkit-font-smoothing: subpixel-antialiased;
	-webkit-user-select: text;
	max-width: 100%;
	overflow-x: scroll;
	pointer-events: auto;
	word-wrap: normal;
}
`
