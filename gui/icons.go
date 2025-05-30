package gui

// Icons thanks to KDE
var kdeIcons = map[string]string{
	"apport": `
<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" xmlns:cc="http://creativecommons.org/ns#" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#" xmlns:inkscape="http://www.inkscape.org/namespaces/inkscape" xmlns:sodipodi="http://sodipodi.sourceforge.net/DTD/sodipodi-0.dtd" width="24" height="24" viewBox="0 0 24 24">
	<style type="text/css" id="current-color-scheme">
		.ColorScheme-Text { color: %s; }
	</style>
	<g transform="translate(1,1)">
		<g id="22-22-apport" transform="translate(0,-10)">
			<path id="path7" class="ColorScheme-Text" d="M 3 13 L 3 29 L 4 29 L 19 29 L 19 28 L 19 13 L 18 13 L 4 13 L 3 13 z M 4 14 L 18 14 L 18 28 L 4 28 L 4 14 z M 5 15 L 5 16 L 6 16 L 6 15 L 5 15 z M 6 16 L 6 18 L 8 18 L 8 16 L 6 16 z M 8 16 L 9 16 L 9 15 L 8 15 L 8 16 z M 8 18 L 8 19 L 9 19 L 9 18 L 8 18 z M 6 18 L 5 18 L 5 19 L 6 19 L 6 18 z M 13 15 L 13 16 L 14 16 L 14 15 L 13 15 z M 14 16 L 14 18 L 16 18 L 16 16 L 14 16 z M 16 16 L 17 16 L 17 15 L 16 15 L 16 16 z M 16 18 L 16 19 L 17 19 L 17 18 L 16 18 z M 14 18 L 13 18 L 13 19 L 14 19 L 14 18 z M 9 21 L 9 23 L 13 23 L 13 21 L 9 21 z M 13 23 L 13 25 L 15 25 L 15 23 L 13 23 z M 15 25 L 15 27 L 17 27 L 17 25 L 15 25 z M 9 23 L 7 23 L 7 25 L 9 25 L 9 23 z M 7 25 L 5 25 L 5 27 L 7 27 L 7 25 z " style="fill:currentColor;fill-opacity:1;stroke:none"/>
			<path id="path9" d="M 0 10 L 0 32 L 22 32 L 22 10 L 0 10 z " style="opacity:1;fill:none"/>
		</g>
	</g>
</svg>
`,

	"checkbox": `
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16">
	<defs id="defs3051">
		<style type="text/css" id="current-color-scheme">
			.ColorScheme-Text { color: %s; }
		</style>
	</defs>
	<path style="fill:currentColor;fill-opacity:1;stroke:none" d="M 13.273438 3.5 L 5.6367188 11.060547 L 2.7265625 8.1796875 L 2 8.9003906 L 4.9082031 11.779297 L 5.6367188 12.5 L 6.7265625 11.419922 L 14 4.2207031 L 13.273438 3.5 z " class="ColorScheme-Text" />
</svg>
`,

	"edit-copy": `
<svg viewBox="0 0 16 16" version="1.1" xmlns="http://www.w3.org/2000/svg">
	<defs>
		<style type="text/css" id="current-color-scheme">
			.ColorScheme-Text { color: %s; }
		</style>
	</defs>
	<path class="ColorScheme-Text" style="fill:currentColor; fill-opacity:1; stroke:none" d="M 3 2 L 3 12 L 6 12 L 6 14 L 14 14 L 14 7 L 11 4 L 10 4 L 8 2 L 3 2 Z M 4 3 L 7 3 L 7 4 L 6 4 L 6 11 L 4 11 L 4 3 Z M 7 5 L 10 5 L 10 8 L 13 8 L 13 13 L 7 13 L 7 5 Z"/>
</svg>
`,

	"go-next": `
<svg viewBox="0 0 16 16" xmlns="http://www.w3.org/2000/svg">
	<style type="text/css" id="current-color-scheme">
		.ColorScheme-Text { color: %s; }
	</style>
	<path d="M11.707 8l-6 6L5 13.293 10.293 8 5 2.707 5.707 2l6 6z" class="ColorScheme-Text" fill="currentColor"/>
</svg>
`,

	"go-previous": `
<svg viewBox="0 0 16 16" xmlns="http://www.w3.org/2000/svg">
	<style type="text/css" id="current-color-scheme">
	.ColorScheme-Text { color: %s; }
	</style>
	<path d="M4.293 8l6 6 .707-.707L5.707 8 11 2.707 10.293 2l-6 6z" class="ColorScheme-Text" fill="currentColor"/>
</svg>
`,

	"view-refresh": `
<svg version="1.1" viewBox="0 0 16 16" xmlns="http://www.w3.org/2000/svg">
	<defs>
		<style type="text/css" id="current-color-scheme">
			.ColorScheme-Text { color: %s; }
		</style>
	</defs>
	<path d="m8 2a6 6 0 0 0-5.9082031 5h1.0234375a5 5 0 0 1 4.8847656-4 5 5 0 0 1 4.564453 3h-3.564453v1h3.896484 0.103516 0.914062 0.085938v-5h-1v2.6894531a6 6 0 0 0-5-2.6894531zm-6 7v5h1v-2.699219a6 6 0 0 0 5 2.699219 6 6 0 0 0 5.908203-5h-1.023437a5 5 0 0 1-4.884766 4 5 5 0 0 1-4.5546875-3h3.5546875v-1h-3.8847656-0.1152344-0.9082031-0.0917969z" class="ColorScheme-Text" fill="currentColor"/>
</svg>
`,
}
