package main

import (
    "bufio"
    "fmt"
    "github.com/hajimehoshi/ebiten"
    "github.com/hajimehoshi/ebiten/ebitenutil"
	"log"
	"math"
	"math/rand"
    "os"
    "strconv"
    "time"
)

// Let's have variables to define the screen window of the game
var screenSizeX = 420
var screenSizeY = 360

// Let's have a variable to define the size of a single game block (cell)
var blockSize = 15

type GameInfo struct {
    level int
    score int
    numEnemies int
    maze []string
}

// Structure to hold information about a single game object
type Sprite struct {
    img *ebiten.Image
	visibility bool
	active bool
	x float64
	y float64
	speed float64
	direction byte
}

// Variable to hold Game Info
var gameInfo GameInfo

// Variable to hold the main game object, THE PACMAN!!!
var pacman Sprite

// Variable to hold maze wall square pieces
var mazeWall []*Sprite

// Variable to hold food
var food []*Sprite

// Variable to keep references to the still objects (wall and food) in order to fast access them. Note that this is a multi-dimensional array having same shape as the maze
var mazeSprites [][]*Sprite

// Variable to hold the enemies
var enemies []*Sprite

// Note that all the arrays above are initialized with * (pointers) to keep only the reference. Otherwise a copy of the object will be created when accessing elements inside them

/*
    Function: createSprite
    Returns a Sprite (Game Object) from a image file.
    Inputs: width and height, initial position (x and y)
*/
func createSprite(imgFile string, width int, height int, x float64, y float64) Sprite {
    // create an empty image with given width and height
    img, _ := ebiten.NewImage(width, height, ebiten.FilterDefault)

    // load pacman image from a file
    imgFromFile, _, err := ebitenutil.NewImageFromFile(imgFile, ebiten.FilterDefault)

    /*
        Let's resize get the size to resize the image according the given height and width
        Ex: image with 600x600 resolution will be resized to given width 15 and height 15 would resize to 15/600= 0.025
    */
    originalWidth, originalHeight := imgFromFile.Size()
    scaleX := float64(width)/float64(originalWidth)
    scaleY := float64(height)/float64(originalHeight)

    // Let's set the resizing to ebiten drawing image options
    opts := &ebiten.DrawImageOptions{}
    opts.GeoM.Scale(scaleX, scaleY)

    // add loaded image to the empty image with resize options
    img.DrawImage(imgFromFile, opts)

    // log if there's any errors occurred while loading the image
	if err != nil {
		log.Fatal(err)
	}

    // return a new Sprite object
	return Sprite{
	    img: img,
	    visibility: true,
	    active: true,
	    x: x,
	    y: y,
	    speed: 1,
	}
}

/*
    Function: setSpritePosition
    Change the XY position value of a given sprite
    Inputs: Reference to the sprite and position (x,y)
    Note that the reference is passed here as we want to update the real Sprite object other than an object copy
*/
func setSpritePosition(sprite *Sprite, x float64, y float64) {
    sprite.x = x
    sprite.y = y
}

/*
    Function: drawSprite
    Render any sprite (Game object) on the screen.
    Inputs: screen and the sprite to render has to be given as arguments
*/
func drawSprite(screen *ebiten.Image, sprite *Sprite) {
    if sprite.visibility {
        opts := &ebiten.DrawImageOptions{}
        opts.GeoM.Translate(sprite.x, sprite.y)
        // opts.GeoM.Scale(sprite.x, sprite.y)
        screen.DrawImage(sprite.img, opts)
    }
}


func readMazeFile(fileName string) []string {
    maze := []string{}
    file, err := os.Open(fileName)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
    	maze = append(maze, line)
    }

    return maze
}
/*
    Function: getPositionFromMaze
    Get the screen position (x, y) from maze character position (row, column)
    Inputs: position of the maze character (row and column)

    Here (int, int) means that the function return two values
*/
func getPositionFromMazePoint(col int, row int) (float64, float64) {
    return float64(blockSize*col), float64(blockSize*row)
}

/*
    Function: getPositionFromMaze
    Get the maze point (column, row) from screen position (x, y)
    Inputs: Screen position (x, y)
*/
func getMazePointFromPosition(x float64, y float64) (int, int) {
    col := int(math.Round(x/float64(blockSize)))
    row := int(math.Round(y/float64(blockSize)))
    return col, row
}

/*
    Function: isValidPoint
    Check if the given point is on the maze or not (boolean)
    Inputs: Maze Point (column, row)
*/
func isValidPoint(col int, row int) bool {
    // Let's define the minimum and maximum column and row values according to the maze matrix
    minRow := 0
    maxRow := len(gameInfo.maze)-1
    minCol := 0
    maxCol := len(gameInfo.maze[0])-1

    if col < minCol || col > maxCol {
        return false
    }

    if row < minRow || row > maxRow {
        return false
    }

    return true
}

/*
    Function: locateGameObjects
    Create game objects (Sprites) and locate them according to the maze (loaded from the file)

    Each character meaning in the maze:
    P - location of the player
    0 - location of a wall piece
    . - Location of a food piece (PacMan can move only through dots)
    E - Enemy which eats the PacMan

*/
func locateGameObjects() {
    // initialize the variable to store pieces of wall with an empty array
    mazeWall = []*Sprite{}

    // initialize the variable to store enemies with an empty array
    enemies = []*Sprite{}

    // initialize the variable to store enemies with an empty array
    food = []*Sprite{}

    // initialize multi dimensional array to hold still element references
    mazeSprites = make([][]*Sprite, len(gameInfo.maze))
    for i := range mazeSprites {
        mazeSprites[i] = make([]*Sprite, len(gameInfo.maze[0]))
    }

    // Read maze which is loaded from the file. each row has a string (line)
    for row, line := range gameInfo.maze {
        // each character in the string (line) is treated as a column
		for col, char := range line {
		    x, y := getPositionFromMazePoint(col, row)
		    // Let's check each character and place corresponding objects to that places
			switch char {
            case '0':
                wall := createSprite("assets/wall.png", blockSize, blockSize, x, y)
                mazeWall = append(mazeWall, &wall)
                mazeSprites[row][col] = &wall
			case 'P':
				pacman = createSprite("assets/pacman.png", blockSize, blockSize, x, y)
				fmt.Printf("PacMan loaded to position: x=%v, y=%v\n", x, y)
            case '.':
                dot := createSprite("assets/food.png", blockSize, blockSize, x, y)
                food = append(food, &dot)
                mazeSprites[row][col] = &dot
			}
		}
	}

	// Now, let's place enemies on random places (random places where there's a path (food))
	for i := 0; i < gameInfo.numEnemies; i++ {
	    // get random food
	    rand.Seed(time.Now().UnixNano())
	    randomFood := food[rand.Intn(len(food))]

	    enemy := createSprite("assets/enemy.png", blockSize, blockSize, randomFood.x, randomFood.y)
        enemies = append(enemies, &enemy)
        fmt.Printf("Enemy loaded to position: x=%v, y=%v\n", randomFood.x, randomFood.y)
	}
}

/*
    Function: movePacman
    Move the pacman on keypress, otherwise keep him idle
*/
func movePacman() {
    /*
        Using ebiten.IsKeyPressed we can check if the given key is pressed at the time of calling this function
        In below, we are using if else because, we want to make sure that only a single key is functional at a given time.
        If none of the below conditions satisfied, pacman will be idle in the current position
    */
    x := pacman.x
    y := pacman.y
    direction := pacman.direction

    // let's get the aligned x and y values to the current location
    col, row := getMazePointFromPosition(x, y)
    alignedX, alignedY := getPositionFromMazePoint(col, row)

    if ebiten.IsKeyPressed(ebiten.KeyUp) {
        // When the "up arrow key" is pressed, let's move the pacman towards north direction from the current position
        y = y-pacman.speed
        x = alignedX
        direction = 'U'
    } else if ebiten.IsKeyPressed(ebiten.KeyDown) {
        // When the "down arrow key" is pressed, let's move the pacman towards south direction from the current position
        y = y+pacman.speed
        x = alignedX
        direction = 'D'
    } else if ebiten.IsKeyPressed(ebiten.KeyLeft) {
        // When the "left arrow key" is pressed, let's move the pacman towards west direction from the current position
        x = x-pacman.speed
        y = alignedY
        direction = 'L'
    } else if ebiten.IsKeyPressed(ebiten.KeyRight) {
        // When the "right arrow key" is pressed, let's move the pacman towards east direction from the current position
        x = x+pacman.speed
        y = alignedY
        direction = 'R'
    } else {
        // PacMan is in idle state. Let's make him to reposition to be in a block (maze point).
        x = alignedX
        y = alignedY
        direction = 'I'
    }

    // Now let's check whether if the new position of PacMan is hitting a Wall
    colNew, rowNew := getMazePointFromPosition(x, y)
    if isValidPoint(colNew, rowNew) && gameInfo.maze[rowNew][colNew] != '0' {
        pacman.x = x
        pacman.y = y
        pacman.direction = direction
    }
}

/*
    Function: eatFood
    Let PacMan eat food if he's on or passing a food sprite
*/
func eatFood() {
    // Let's get the current position of the pacman to map to the maze point
    col, row := getMazePointFromPosition(pacman.x, pacman.y)

    // check the symbol at that point in the maze matching food symbol (i.e. dot)
    if isValidPoint(col, row) && gameInfo.maze[row][col] == '.' {
        // player is on a food, remove the . from maze
        gameInfo.maze[row] = gameInfo.maze[row][:col] + " " + gameInfo.maze[row][col+1:]

        // make the food invisible from the screen
        mazeSprites[row][col].visibility = false

        // increase the player score by 1
        gameInfo.score = gameInfo.score+1
    }
}

// RE-WRITE THIS FUNCTION
func moveEnemy(sprite *Sprite) {
    x := sprite.x
    y := sprite.y
    // direction := sprite.direction

    // current maze point of the enemy
    col, row := getMazePointFromPosition(x, y)

    // Let's find a good direction for the enemy to move

    // To find that let's have a scoring mechanism to each direction which Enemy can move
    movementScore := [4]int{0, 0, 0, 0} // array indices are in order UP, RIGHT , DOWN, LEFT

    // grant 1 point to UP if UP is a valid point and no wall
    if isValidPoint(col, row-1) &&  gameInfo.maze[row-1][col] != '0' {
        movementScore[0] = movementScore[0]+1
    }
    // grant 1 point to RIGHT if RIGHT is a valid point and no wall
    if isValidPoint(col+1, row) &&  gameInfo.maze[row][col+1] != '0' {
        movementScore[1] = movementScore[1]+1
    }
    // grant 1 point to DOWN if DOWN is a valid point and no wall
    if isValidPoint(col, row+1) &&  gameInfo.maze[row+1][col] != '0' {
        movementScore[2] = movementScore[2]+1
    }
    // grant 1 point to LEFT if LEFT is a valid point and no wall
    if isValidPoint(col-1, row) &&  gameInfo.maze[row][col-1] != '0' {
        movementScore[3] = movementScore[3]+1
    }

    // Let's find the directions with maximum scores
    maxScores := []int{}
    for i, score := range movementScore {
        if len(maxScores) == 0 {
            maxScores = append(maxScores, i)
        } else if (score == maxScores[0]) {
            maxScores = append(maxScores, i)
        } else if (score > maxScores[0]) {
            maxScores = []int{i}
        }
    }

    // out of maximums, lets get a random direction
    movingDirectionIndex := rand.Intn(len(maxScores))

    switch maxScores[movingDirectionIndex] {
    case 0:
        sprite.y = sprite.y-sprite.speed
    case 1:
        sprite.x = sprite.x+sprite.speed
    case 2:
        sprite.y = sprite.y+sprite.speed
    case 3:
        sprite.x = sprite.x-sprite.speed
    }

    // fmt.Printf("Max total: %v direction: %v\n", len(maxScores), maxScores[movingDirectionIndex])

}

// code inside update function is called every 60 times per second
func update(screen *ebiten.Image) error {
    // Let's skip rendering the frame is the game play gets slow. (This is increases the performance)
	if ebiten.IsDrawingSkipped() {
	    // stop the function here
		return nil
	}

	// Let's code what should happen on each frame (Game Starts from here)
    movePacman()
    eatFood()

	for _, wallPiece := range mazeWall {
	    drawSprite(screen, wallPiece)
    }

    for _, dot := range food {
	    drawSprite(screen, dot)
    }

    for _, enemy := range enemies {
        moveEnemy(enemy)
	    drawSprite(screen, enemy)
    }

    drawSprite(screen, &pacman)

    ebitenutil.DebugPrint(screen, "Score: "+strconv.Itoa(gameInfo.score))
	return nil
}

func main() {
    gameInfo = GameInfo {
        level: 1,
        score: 1,
        numEnemies: 4,
        maze: readMazeFile("maze02.txt"),
    }

    locateGameObjects()

	if err := ebiten.Run(update, screenSizeX, screenSizeY, 2, "Simple PacMan Game"); err != nil {
		log.Fatal(err)
	}
}
