# Instagram Media Downloader Bot (Python & Go Versions)

A Telegram bot for downloading media from Instagram and other websites.  
**Live Bot:** [@igramloadbot](https://t.me/igramloadbot)

## Features

- Download public Instagram posts (videos and photos)
- Download content from other supported websites via yt-dlp
- Available in both Python and Go implementations
- Both direct messages and inline queries supported

## Python Version

### Requirements
- Python 3.8+
- yt-dlp
- python-telegram-bot

### Installation
```bash
pip install -r requirements.txt
```

### Running
```bash
python bot.py
```

## Go Version

### Requirements
- Go 1.20+
- yt-dlp (via Python requirements.txt)

### Installation
```bash
go mod download
pip install -r requirements.txt  # for yt-dlp
```

### Running
```bash
go run scraper.go
```

## Common Configuration
```json
// secrets.json
{
  "bot_token": "YOUR_TELEGRAM_BOT_TOKEN", 
  "chat_id": YOUR_CHAT_ID
}
```

## Limitations
⚠️ **Important Notes:**
- Only works with **public** content
- Private/restricted posts will show error message
- Instagram Reels/Stories may not always work
- Download speed depends on internet connection

## Supported Websites
The bot can download from:
- Instagram (public posts)
- YouTube, Twitter, TikTok, Facebook
- [Full list of supported sites](https://github.com/yt-dlp/yt-dlp/blob/master/supportedsites.md)

## Troubleshooting
If downloads fail:
1. Verify content is public
2. Check internet connection  
3. Update yt-dlp: `pip install -U yt-dlp`
4. Check logs for error details

## License
MIT
