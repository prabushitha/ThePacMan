## Run the Game
- Run the command `./SimplePacmanGame` from the terminal

## Instructions to build and run

- Install Go programming language
- Install ebiten package (https://ebiten.org/install.html)
- run the command `go build` in the terminal (This will create an executable)
- run the command `./SimplePacmanGame` in the terminal

## Source code
To understand the things easily I've written the complete game in a single file `main.go`
Refer the comments I have made to understand the code.

### Window, the Grid and Game Objects
We are considering the screen as a grid. Size of each cell in the grid is defined by the global variable `blockSize`.
All the game elements (images) except "GAME OVER", "LEVEL UP", "YOU WIN" and "START" are size of the block.

Block Size elements: `Pacman, Enemy and Wall`

Find the png file `HowGridWorks.png` to get an understanding about the grid and positioning.

## Reference for Logos and Graphics 
pacman - https://pngriver.com/download-pac-man-png-clipart-74647/

enemies - http://www.pngmart.com/files/2/Pac-Man-Ghost-PNG-Photos.png

food - https://t5.rbxcdn.com/4db7f1bd2c5e6c439edf891b595e30ce

wall - https://s1.construct.net/images/v711/uploads/articleuploadobject/0/images/13274/chompermazetiles.png

win popup - https://s.pngix.com/pngfile/s/149-1496672_you-win-you-win-pixel-art-hd-png.png

game over popup - https://myrealdomain.com/images/game-over-png-4.png

levelup popup - http://www.sclance.com/pngs/level-up-png/level_up_png_787135.png

start popup - https://www.pngfind.com/pngs/m/37-374698_pac-man-logos-01-by-dhlarson-d5qqh82-2.png