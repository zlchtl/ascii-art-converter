# Image to ASCII Art Converter

This project provides a service that converts images to ASCII art. The service can be accessed via a REST API, where you send an image and specify parameters like the size and character set used to create the ASCII art.

## Components

### 1. Constants

- **DEFAULTSIZE**: Default image size for resizing, set to 100.
- **ASCII_CHARS**: The string of characters used for ASCII art generation. Default is "@%#*+=-:. ".
- **MAX_IMAGE_SIZE**: The maximum image size that can be processed (8 MB).

### 2. Data Structures

##### ConvertParams

This structure contains the parameters used for converting the image:
```go
type ConvertParams struct {
    Size    int    `json:"size" binding:"gte=1,lte=300"`
    CharSet string `json:"charSet" binding:"required,min=2,max=32"`
}
```
- **Size** (int): The size of the final image (1 ≤ size ≤ 300).
- **CharSet** (string): The character set to use for the ASCII art (2 to 32 characters).

### 3. HTTP Handlers

##### `convertHandler(c *gin.Context)`

Handles POST requests to convert an image to ASCII art. You need to provide an image and parameters in the request body.

**Request Parameters**:
- `file`: The image file to be converted (JPEG or PNG format).
- `params`: A JSON object containing the conversion parameters:
  - `size`: The final image size (1 ≤ size ≤ 300).
  - `charSet`: The set of characters to be used for the ASCII art (2 to 32 characters).

**Example Request**:
```bash
curl -X POST http://localhost:8080/convert
-F "file=@images.jpg"
-F "params={"size":150, "charSet":"@%#*+=-:. "}"
```

**Response**:
- On success, the ASCII art will be returned in a JSON response:
```bash
{"ascii": "ASCII_ART"}
```

- On error, a detailed error message will be returned:
```bash
{"error": "Error message"}
```

##### `aboutHandler(c *gin.Context)`

Handles GET requests to provide information about the project.

**Response**:
```bash
{ 
	"about": "Project of image translation to ASCI code", 
	"tip": "For the algorithm to work more correctly, contrasting images are required." 
}
```

### 4. Image Processing Functions

##### `resizeImage(img image.Image, size int) image.Image`

Resizes the image to the specified size while preserving the aspect ratio.

**Parameters**:
- `img`: The original image.
- `size`: The new size for the image (either width or height).

**Returns**:
- A resized image.

##### `convertToGray(img image.Image) *image.Gray`

Converts an image to grayscale.

**Parameters**:
- `img`: The original image.

**Returns**:
- A grayscale image.

##### `convertToASCII(ctx context.Context, file io.Reader, filename string, params ConvertParams) (string, error)`

Converts the image to ASCII art.

**Parameters**:
- `ctx`: The context for managing operation timeouts.
- `file`: The image file to be processed.
- `filename`: The name of the image file (used to determine the format).
- `params`: The conversion parameters (size and character set).

**Returns**:
- A string containing the ASCII art.

##### `generateASCII(img *image.Gray, charset string) string`

Generates the ASCII art string from a grayscale image.

**Parameters**:
- `img`: The grayscale image.
- `charset`: The character set used for displaying pixel brightness.

**Returns**:
- A string containing the ASCII art.

### 5. Running the Server

To run the server, use the following command:
```bash
go run main.go
```

The server will be available at `http://localhost:8080`.

### 6. Error Handling

Possible errors:
- **400 Bad Request**: An issue with the file upload (e.g., no file provided or unsupported format).
- **404 Not Found**: Invalid path or HTTP method.
- **500 Internal Server Error**: An error on the server (e.g., failure to process the image).
- **504 Gateway Timeout**: The operation timed out while processing the image.