# simplelinkshortener

simplelinkshortener is a standalone application for creating private self-hosted link shortening services.

## Usage

### 1. Install the latest version:
```bash
go install "github.com/salmanmorshed/simplelinkshortener@latest"
```

### 2. Run the init command:
```bash
~/go/bin/simplelinkshortener init
```
This command will guide you through the initial setup and generate a config file containing database and web server configuration. It'll also generate a randomized alphabet required to create the short links. You can specify the location of the config file using the global `--config` option.

### 3. Start the webserver:
```bash
~/go/bin/simplelinkshortener start
```
The server will begin listening to web traffic. Users with valid credentials can access the service to create and access short links.

## User Management
Access to the API and the web UI is restricted by HTTP Basic Authentication. You must create user accounts before using the shortener. To add a new user, use the `useradd` command. Check the help menu for more details: 
```bash
~/go/bin/simplelinkshortener --help
```


## API Endpoints
### 1. Create a new short link

- **URL**: `/api/links`
- **Method**: POST
- **Authentication**: Basic Authentication
- **Request Body**: JSON with a `url` field (string, required).
- **Response**: Shortened URL as `short_url`.

**Example Request:**
```http
POST /api/links
Content-Type: application/json

{
  "url": "https://example.com"
}
```

**Response:**
```http
HTTP/1.1 201 Created
Content-Type: application/json

{
  "short_url": "https://short.dev/abcde"
}
```

### 2. Retrieve links created by a user

- **URL**: `/api/links`
- **Method**: GET
- **Authentication**: Basic Authentication
- **Query Parameters**: 
  - `limit` (integer) default = 10
  - `offset` (integer) default = 0
- **Response**: List of user's short links with pagination details.

**Example Request:**
```http
GET /api/links

HTTP/1.1 200 OK
Content-Type: application/json
```

**Response:**
```json
{
  "results": [
    {
      "id": "abcde",
      "url": "https://example.com",
      "visits": 5,
      "created_at": "2023-04-20T06:09:00Z"
    }
  ],
  "total": 1,
  "limit": 10,
  "offset": 0,
  "prefix": "https://short.dev"
}
```

## Web frontend
A work-in-progress frontend app is served on `/web`. You can use it to create or view your links.


## License
This project is licensed under the [MIT License](https://github.com/git/git-scm.com/blob/main/MIT-LICENSE.txt). The MIT License is a permissive open-source license that allows you to freely use, modify, and distribute this software for both commercial and non-commercial purposes, provided you include the original copyright notice and disclaimer. Feel free to explore, contribute, and build upon this project with confidence under the terms of the MIT License.
