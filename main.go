package main

/*
Let's import the required packages
Except "github.com/hajimehoshi/ebiten", all the other packages are default packages that comes with Go Language
*/
import (
    "bufio"
    "github.com/hajimehoshi/ebiten"
    "github.com/hajimehoshi/ebiten/ebitenutil"
	"log"
	"math"
	"math/rand"
    "os"
    "strconv"
    "time"
)

/*
    ################
    ## Structures ##
    ################
*/
// Structure which keeps information about the current game play
type GameInfo struct {
    level int // holds current level
    maxScore int // holds maximum score to achieve inorder to finish the level (number of food points)
    score int // holds score of current level (number of food eaten by PacMan)
    isStarted bool // when the game is started (PacMan is moving), this flag is set to true
    isGameOver bool // when the game is over (enemy eat PacMan), this flag is set to true
    isLevelComplete bool // when the level is completed (PacMan eat all food), this flag is set to true
    maze []string // holds the maze file as string array, each string is a row. each character in the string is a column
}

// Structure which keeps information about a level
type LevelInfo struct {
    pacmanSpeed float64 // holds the speed of the PacMan on a level
    enemySpeed float64 // holds the speed of an enemy on a level
    numEnemies int // holds the number of elements which should loaded into a level
    mazeFile string // holds the path to the file containing the maze on a level
}

// Structure to hold information about a single game object
type Sprite struct {
    img *ebiten.Image // holds the image displayed as the game object
    faces map[byte]*ebiten.Image // holds the images of different faces/animations of this sprite. This is a map structure
	visibility bool // holds if the game object is visible or not
	x float64 // holds the x position of the game object in the screen
	y float64 // holds the y position of the game object in the screen
	speed float64 // holds the speed of moving game objects (used for PacMan and enemies)
	direction byte // holds the current moving direction of moving game objects (U=UP, R=RIGHT, D=DOWN, L=LEFT)
}

/*
    ###############################
    ## Defining Global Variables ##
    ###############################
*/

// Let's have variables to hold the Size of screen window of the game
var screenSizeX = 420
var screenSizeY = 360

// Let's have a variable to define the size of a single game block (cell)
var blockSize = 15

// Let's define all the levels for the game
var LEVELS = map[int]LevelInfo {
    1: LevelInfo{
        pacmanSpeed: 2,
        enemySpeed: 2,
        numEnemies: 4,
        mazeFile: "maze01.txt",
    },
    2: LevelInfo{
        pacmanSpeed: 2,
        enemySpeed: 3,
        numEnemies: 5,
        mazeFile: "maze02.txt",
    },
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
    Variables to hold popups
    Popups in the game are: Game Over, Level Complete, Win, Start
*/
var gameOver Sprite
var levelComplete Sprite
var win Sprite
var startLogo Sprite

/*
    ###############################################
    ## Functions to read files (maze and assets) ##
    ###############################################
*/

/*
    Function: readMazeFile
    Read file which containing the maze information
    Inputs: path to the file
    Outputs an array containing the rows of the maze, each row as a string
*/
func readMazeFile(fileName string) []string {
    // create an empty array to hold the maze
    maze := []string{}

    // Open the file and load bytes into a variable
    file, err := os.Open(fileName)

    // if error occurred while loading the file log it
    if err != nil {
        log.Fatal(err)
    }
    // close the file once this method has completely executed
    defer file.Close()

    // create a scanner to read the bytes as string
    scanner := bufio.NewScanner(file)

    // get rows from the scanner until all rows has finished scanning
    for scanner.Scan() {
        // get the row as a string
        line := scanner.Text()

        // push each string line to the maze array
    	maze = append(maze, line)
    }

    return maze
}

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
	    x: x,
	    y: y,
	    speed: 1,
	}
}


/*
    ##########################################
    ## Position and Grid supporting methods ##
    ##########################################
*/

/*
    Function: getPositionFromMaze
    Get the screen position (x, y) from maze character position (row, column)
    Inputs: position of the maze character (row and column)

    Here (int, int) means that the function return two values
*/
func getPositionFromMazePoint(col int, row int) (float64, float64) {
    // here we are casting to float as we standard position usage in ebiten is float
    return float64(blockSize*col), float64(blockSize*row)
}

/*
    Function: getPositionFromMaze
    Get the maze point (column, row) from screen position (x, y)
    Inputs: Screen position (x, y)
*/
func getMazePointFromPosition(x float64, y float64) (int, int) {
    /*
    We have to find the correct grid cell when the screen position (x, y) is supplied
        grid column at x = x/blockSize (integer value)
        grid column at y = y/blockSize (integer value)
    */
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
    minRow := 0 // 1st row
    maxRow := len(gameInfo.maze)-1 // last row

    minCol := 0 // 1st column
    maxCol := len(gameInfo.maze[0])-1 // last column

    // if given column is not within the 1st column and last column, the point is invalid
    if col < minCol || col > maxCol {
        return false
    }

    // if given row is not within the 1st row and last row, the point is invalid
    if row < minRow || row > maxRow {
        return false
    }

    // otherwise it's a valid point
    return true
}

/*
    Function: getMovableDirection
    Get a movable direction from the given maze point
    Outputs a byte indicating direction: U=UP, R=RIGHT , D=DOWN, L=LEFT
*/
func getMovableDirection(col int, row int, currentDirection byte) byte {
    // To find that let's have a array to store all possible directions
    possibilities := []byte{}

    // let's also have an map to store if any direction is possible
    directions := map[byte]bool{
        'U': false,
        'R': false,
        'D': false,
        'L': false,
    }

    // add UP if UP is a valid point and no wall
    if isValidPoint(col, row-1) &&  gameInfo.maze[row-1][col] != '0' {
        possibilities = append(possibilities, 'U')
        directions['U']=true
    }
    // add RIGHT if RIGHT is a valid point and no wall
    if isValidPoint(col+1, row) &&  gameInfo.maze[row][col+1] != '0' {
        possibilities = append(possibilities, 'R')
        directions['R']=true
    }
    // add DOWN if DOWN is a valid point and no wall
    if isValidPoint(col, row+1) &&  gameInfo.maze[row+1][col] != '0' {
        possibilities = append(possibilities, 'D')
        directions['D']=true
    }
    // add LEFT if LEFT is a valid point and no wall
    if isValidPoint(col-1, row) &&  gameInfo.maze[row][col-1] != '0' {
        possibilities = append(possibilities, 'L')
        directions['L']=true
    }

    // Let's get a random direction out of all the possible directions. This is the standard way of generating a random integer in Go.
    rand.Seed(time.Now().UnixNano())
    direction := possibilities[rand.Intn(len(possibilities))]

    // if the direction we get is UP but sprite is moving DOWN and still possible to move DOWN, move it DOWN!
    if direction == 'U' && currentDirection == 'D' && directions['D'] {
        return 'D'
    }

    // if the direction we get is DOWN but sprite is moving UP and still possible to move UP, move it UP!
    if direction == 'D' && currentDirection == 'U' && directions['U'] {
        return 'U'
    }

    // if the direction we get is LEFT but sprite is moving RIGHT and still possible to move RIGHT, move it RIGHT!
    if direction == 'L' && currentDirection == 'R' && directions['R'] {
        return 'R'
    }

    // if the direction we get is LEFT but sprite is moving RIGHT and still possible to move RIGHT, move it RIGHT!
    if direction == 'R' && currentDirection == 'L' && directions['L'] {
        return 'L'
    }

    // if it doesn't satisfy above conditions, let's return the direction we got!
    return direction
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

/*
    ############################################
    ## Defining behaviours of movable objects ##
    ############################################
*/

/*
    Function: movePacman
    Move the PacMan on keypress, otherwise keep him idle
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

    // let's get the aligned x and y values to the current location (aligned values means the values which makes PacMan center on the path)
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
        // it's a valid point in the maze and there's no wall in this point. PacMan is good to move. Let's move it to the new position
        pacman.x = x
        pacman.y = y
        pacman.direction = direction

        // Now let's set the face of the PacMan according to the direction
        pacman.img = pacman.faces[direction]
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

    // let's check if user has eat all food. if all food has been eaten, let's complete the level
    if gameInfo.score >= gameInfo.maxScore {
        gameInfo.isLevelComplete = true
    }
}

/*
    Function: moveEnemy
    Moving a given enemy for a possible direction
    Input: reference to a enemy game object
*/
func moveEnemy(sprite *Sprite) {
    x := sprite.x
    y := sprite.y

    // current maze point of the enemy
    col, row := getMazePointFromPosition(x, y)

    // current maze point of the pacman
    colPac, rowPac := getMazePointFromPosition(pacman.x, pacman.y)

    // Let's check if ENEMIE HIT the PACMAN!. If so make game over
    if col == colPac && row == rowPac{
        gameInfo.isGameOver = true
    }

    // Let's get the aligned position to keep enemy on center of the path
    alignedX, alignedY := getPositionFromMazePoint(col, row)

    // make the direction to point the direction where enemy is currently moving
    direction := sprite.direction

    /*
        Enemy should not move in a single direction always. We need to find a movable direction at a junction point.
        You can have infinite number of positions (x,y values) in a grid cell.
        Therefore, if enemy check for possible directions at all positions in a junction, he may try to vibrate at junction.
        Reason is that at each position in a junction, enemy thinks it's a new junction and tries to find a new direction to move.
        So ideally, this should not happen.

        Let's do a small trick to get rid of the above scenario.
        In a junction, let's check for the distance between the grid cell left corner position and enemy's left corner position.
        If the enemy has moved reasonable amount (identifiable amount which can make enemy to be in next block in the next move) of distance from the current grid only we find for a new direction.
    */
    reasonableMoveAmount := math.Floor(float64(blockSize)/2.0)-1.0 // This equation has been taken on trial and error basis. if the block size is 15, reasonable amount is 6.
    if math.Abs(x-alignedX) > reasonableMoveAmount || math.Abs(y-alignedY) > reasonableMoveAmount {
        // get a movable direction
        direction = getMovableDirection(col, row, sprite.direction)
    }
    sprite.direction = direction

    // Let's move the enemy
    switch direction {
    case 'U':
        sprite.y = sprite.y-sprite.speed
        sprite.x = alignedX
    case 'R':
        sprite.x = sprite.x+sprite.speed
        sprite.y = alignedY
    case 'D':
        sprite.y = sprite.y+sprite.speed
        sprite.x = alignedX
    case 'L':
        sprite.x = sprite.x-sprite.speed
        sprite.y = alignedY
    }
}


/*
    ########################################
    ## Functions to initialize properties ##
    ##      and assets for the game       ##
    ########################################
*/

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
		    // let's get the position (left corner position x, y) of the grid cell
		    x, y := getPositionFromMazePoint(col, row)
		    // Let's check each character and place corresponding objects to that places
			switch char {
            case '0':
                // create a Wall block for 0 point in maze and mark position to the corresponding grid cell
                wall := createSprite("assets/wall.png", blockSize, blockSize, x, y)
                // let's store the wall block Sprite reference in the mazeWall array
                mazeWall = append(mazeWall, &wall)

                // let's store the wall block Sprite reference in the maze grid matrix as well to quickly get the object
                mazeSprites[row][col] = &wall
			case 'P':
			    // create the PacMan and mark position to the corresponding grid cell
				pacman = createSprite("assets/pacman.png", blockSize, blockSize, x, y)

				// now let's load the other faces of pacman
				UP_SPRITE := createSprite("assets/pacmanU.png", blockSize, blockSize, x, y)
				RIGHT_SPRITE := createSprite("assets/pacmanR.png", blockSize, blockSize, x, y)
				DOWN_SPRITE := createSprite("assets/pacmanD.png", blockSize, blockSize, x, y)
				LEFT_SPRITE := createSprite("assets/pacmanL.png", blockSize, blockSize, x, y)
				IDLE_SPRITE := createSprite("assets/pacmanI.png", blockSize, blockSize, x, y)

                pacman.faces = map[byte]*ebiten.Image{
                    'U': UP_SPRITE.img,
                    'R': RIGHT_SPRITE.img,
                    'D': DOWN_SPRITE.img,
                    'L': LEFT_SPRITE.img,
                    'I': IDLE_SPRITE.img,
                }

				// Since PacMan is moving always, we don't need to add it to the maze grid matrix
            case '.':
                // create the food and mark position to the corresponding grid cell
                dot := createSprite("assets/food.png", blockSize, blockSize, x, y)

                // let's store the food block Sprite reference in the mazeWall array
                food = append(food, &dot)

                // let's store the food block Sprite reference in the maze grid matrix as well to quickly get the object
                mazeSprites[row][col] = &dot

                // foods are our scores, it's increase the maximum possible score by 1 as food is added to the maze
                gameInfo.maxScore = gameInfo.maxScore+1
			}
		}
	}

	// Now, let's place enemies on random places (random places where there's a path (food))
	for i := 0; i < LEVELS[gameInfo.level].numEnemies; i++ {
	    // get random food. This is the standard way of generating a random integer in Go.
	    rand.Seed(time.Now().UnixNano())
	    randomFood := food[rand.Intn(len(food))]

        // Let's create and enemy and mark its location at the random food. This way we can place enemies at random points in a movable path
	    enemy := createSprite("assets/enemy.png", blockSize, blockSize, randomFood.x, randomFood.y)

        // Let's also give an initial direction for the enemy to move
        // For this we need to get the grid point which this enemy is getting placed
	    colFood, rowFood := getMazePointFromPosition(randomFood.x, randomFood.y)
	    // Now get a possible movable direction at that grid cell
        enemy.direction = getMovableDirection(colFood, rowFood, enemy.direction)

        // Let's add enemy to the list of enemies. We don't need to add to maze grid matrix as enemy is moving.
        enemies = append(enemies, &enemy)
	}

	// Let's load game over image as Sprite
	gameOver = createSprite("assets/gameover.png", blockSize*10, blockSize*7, float64(screenSizeX)/2.0-float64(blockSize*10)/2.0, float64(screenSizeY)/2.0-float64(blockSize*7)/2.0)

	// Let's load level complete image as Sprite
    levelComplete = createSprite("assets/levelcomplete.png", blockSize*12, blockSize*12, float64(screenSizeX)/2.0-float64(blockSize*12)/2, float64(screenSizeY)/2.0-float64(blockSize*12)/2)

    // Let's load Win image as Sprite
    win = createSprite("assets/win.png", blockSize*12, blockSize*12, float64(screenSizeX)/2.0-float64(blockSize*12)/2, float64(screenSizeY)/2.0-float64(blockSize*12)/2)

    // Let's load Start logo image as Sprite
    startLogo = createSprite("assets/start.png", blockSize*14, blockSize*5, float64(screenSizeX)/2.0-float64(blockSize*14)/2, float64(screenSizeY)/2.0-float64(blockSize*5)/2)
}

/*
    Function: initLevel
    Initialize game information to use the given level
    Input: level
*/
func initLevel(level int) {
    // initialize the game info
    gameInfo = GameInfo {
        level: level,
        score: 1,
        maxScore: 1, // this will be set after loading all the food sprites. for now let's keep it as 1
        maze: readMazeFile(LEVELS[level].mazeFile),
    }

    // load game objects from assets and locate them in corresponding places
    locateGameObjects()
}



/*
    ####################
    ## Main Game Loop ##
    ####################
*/

// code inside update function is called every 60 times per second
func update(screen *ebiten.Image) error {
    // Let's skip rendering the frame is the game play gets slow. (This is increases the performance)
	if ebiten.IsDrawingSkipped() {
	    // stop the function here
		return nil
	}

	// Let's code what should happen on each frame (Game Starts from here)

    // Let's draw the Walls and food first
    // Go through all the sprites in mazeWall and add them to the screen
    for _, wallPiece := range mazeWall {
	    drawSprite(screen, wallPiece)
    }

    // Go through all the sprites in food and add them to the screen
    for _, dot := range food {
	    drawSprite(screen, dot)
    }

    if !gameInfo.isStarted {
        // Show Start screen when game is not yet started
        drawSprite(screen, &startLogo)

        // When space is pressed, load next level
        if ebiten.IsKeyPressed(ebiten.KeySpace) {
            // hide start logo complete
            gameInfo.isStarted = true
        }

    } else if gameInfo.isLevelComplete {
        // Show Level Complete / WIN Screen on level complete
        nextLevel := gameInfo.level+1

        // if the Next level is below number of levels, show level complete text
        if nextLevel <= len(LEVELS) {
            drawSprite(screen, &levelComplete)
            _, h := levelComplete.img.Size()
            ebitenutil.DebugPrintAt(screen, "Press Space to START.....", int(levelComplete.x)+blockSize, int(levelComplete.y)+h)
        } else {
            // if the Next level is above number of levels, that means user has completed all the levels. Let's show win screen
            drawSprite(screen, &win)
            // show Text under the win sprite
            _, h := win.img.Size()
            ebitenutil.DebugPrintAt(screen, "Press Space to START..", int(win.x)+2*blockSize, int(win.y)+h)

            // let's make the next level as 1, to start over when space is pressed
            nextLevel = 1
        }

        // When space is pressed, load next level
        if ebiten.IsKeyPressed(ebiten.KeySpace) {
            // hide level complete
            gameInfo.isLevelComplete = false

            // load next level
            initLevel(nextLevel)
        }

    } else if gameInfo.isGameOver {
        // Show Game Over Screen  on game over
        drawSprite(screen, &gameOver)

        _, h := gameOver.img.Size()
        ebitenutil.DebugPrintAt(screen, "Press Space to START", int(gameOver.x)+blockSize, int(gameOver.y)+h)

        // When space is pressed, start from level 1
        if ebiten.IsKeyPressed(ebiten.KeySpace) {
            // hide game over
            gameInfo.isGameOver = false
            // load level 1
            initLevel(1)
        }
    } else {
        // There are no any pause screens, Let's make the pacman and enemies visible and allow them to move

        // Main Game logic exist here
        // Let's move the PacMan if user is pressing a direction key
        movePacman()

        // let PacMan eat food, if there's any food on the current location
        eatFood()

        // get each enemy from the list of enemies array and move each enemy
        for _, enemy := range enemies {
            // move enemy to a possible direction
            moveEnemy(enemy)
            // show enemy on the screen
    	    drawSprite(screen, enemy)
        }

        // show the PACMAN on screen
        drawSprite(screen, &pacman)
    }

    // show the score and level on top left corner of the screen
    ebitenutil.DebugPrint(screen, "  Level: "+strconv.Itoa(gameInfo.level)+"   Score: "+strconv.Itoa(gameInfo.score))
	return nil
}


/*
    #################
    ## MAIN METHOD ##
    #################
*/
// When we run this GO file, this method is getting executed first.
func main() {
    // Let's initialize the game information use level 1
    initLevel(1)

    // ebiten.Run is a function given by the ebiten library.
    // Here, we give a method which should call always (60 times per second) and size of the screen, scale the window by 1.5 and name of the window as Simple PacMan Game
	err := ebiten.Run(update, screenSizeX, screenSizeY, 1.5, "Simple PacMan Game")
	// Note that the update method contain all the game logic

	// If there's any error occured in ebiten library to fail loading the window, let's log it
	if err != nil {
		log.Fatal(err)
	}
}
