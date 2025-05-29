# fontview

A tool to view every damn glyph a font contains, along with plenty of useful Unicode info.

## features

- Fast
- Sorted in order
  - (this is the main reason I started this project)
- Unicode details
- HTML Named entities, like `&int;` for `∫`
- Plenty of copy formats
  - Raise an issue for more formats
- Massive preview
- No updates needed
  - Automatically fetches the latest Unicode & HTML Entity data every month

> [!WARNING]
> This app is only tested on KDE. Do not raise issues about the missing copy icon.

![image](./.media/gopher.png)
![image](./.media/unicode.png)
![image](./.media/html.png)

To do:

- [x] Fast table (do not load 50k lines at once)
- [x] Installed fonts
- [ ] Custom font file
- [ ] Search
- [ ] List glyph name in font
- [ ] Reference history
