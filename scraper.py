import os
import yt_dlp
import time
import threading
import logging
from telegram import InlineQueryResultArticle, Update, InputFile, InlineQueryResultCachedVideo, InlineQueryResultCachedPhoto, InputTextMessageContent
from telegram.ext import Updater, CommandHandler, MessageHandler, Filters, InlineQueryHandler
from collections import defaultdict
from datetime import datetime, timedelta

# Configure logging
logging.basicConfig(
    level=logging.DEBUG,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[
        logging.StreamHandler(),
        logging.FileHandler('scraper.log')
    ]
)
logger = logging.getLogger(__name__)

# Load secrets from JSON file
try:
    import json
    with open('secrets.json') as f:
        secrets = json.load(f)
    BOT_TOKEN = secrets['bot_token']
    CHAT_ID = secrets.get('chat_id', 824640741)  # Default fallback
except Exception as e:
    logger.error(f"Failed to load secrets: {e}")
    raise RuntimeError("Missing required secrets configuration")
DOWNLOADS_DIR = os.path.join(os.path.dirname(__file__), "downloads")
os.makedirs(DOWNLOADS_DIR, exist_ok=True)
CLEANUP_INTERVAL = 30
REQUEST_LIMIT = 15
MAX_CONCURRENT_DOWNLOADS = 30

user_requests = defaultdict(list)
lock = threading.Lock()

def download_instagram_media(url: str, user_id: str):
    """Reliable Instagram media downloader using yt-dlp"""
    logger.info(f"Starting download for {url}")
    user_dir = os.path.join(DOWNLOADS_DIR, str(user_id))
    os.makedirs(user_dir, exist_ok=True)
    timestamp = str(int(time.time()))
    file_path = os.path.join(user_dir, f'{timestamp}_instagram_media')

    ydl_opts = {
        'outtmpl': os.path.join(user_dir, f'{timestamp}_%(title)s.%(ext)s'),
        'quiet': True,
        'socket_timeout': 15,
        'retries': 3,
        'format': 'best',
        'writethumbnail': False,
        'nocheckcertificate': True,
        'extractor_args': {
            'instagram': {
                'extract_flat': True
            }
        }
    }

    try:
        with yt_dlp.YoutubeDL(ydl_opts) as ydl:
            info = ydl.extract_info(url, download=True)
            file_path = ydl.prepare_filename(info)
            
            if not os.path.exists(file_path):
                raise Exception("Download failed - file not created")
            
            is_video = file_path.lower().endswith(('.mp4', '.mkv', '.webm'))
            
            return {
                'path': file_path,
                'is_video': is_video,
                'title': info.get('title', 'Instagram Media')[:64],
                'user_id': user_id
            }
    except Exception as e:
        logger.error(f"Download failed: {str(e)}")
        return None

def check_rate_limit(user_id):
    """Implement rate limiting per user"""
    with lock:
        now = datetime.now()
        user_requests[user_id] = [t for t in user_requests[user_id] if now - t < timedelta(minutes=1)]
        if len(user_requests[user_id]) >= REQUEST_LIMIT:
            return False
        user_requests[user_id].append(now)
        return True

def clean_old_files():
    """Remove files older than CLEANUP_INTERVAL"""
    now = time.time()
    for user_dir in os.listdir(DOWNLOADS_DIR):
        user_path = os.path.join(DOWNLOADS_DIR, user_dir)
        if os.path.isdir(user_path):
            for filename in os.listdir(user_path):
                filepath = os.path.join(user_path, filename)
                if os.path.isfile(filepath):
                    file_time = os.path.getmtime(filepath)
                    if now - file_time > CLEANUP_INTERVAL:
                        try:
                            os.remove(filepath)
                        except Exception as e:
                            logger.error(f"Error cleaning file {filepath}: {e}")

def start(update: Update, context):
    update.message.reply_text("Send me an Instagram/Youtube link, and I'll fetch the media for you!")

def handle_message(update: Update, context):
    user_id = update.message.from_user.id
    url = update.message.text.strip()
    
    if not url.startswith(('http://', 'https://')):
        update.message.reply_text("Please send a valid URL starting with http:// or https://")
        return
    
    if not check_rate_limit(user_id):
        update.message.reply_text("⚠️ Please wait a minute before making more requests")
        return
    
    update.message.reply_text("⏳ Downloading media, please wait...")
    
    result = download_instagram_media(url, user_id)
    if result:
        try:
            with open(result['path'], "rb") as media_file:
                if result['is_video']:
                    update.message.reply_video(
                        video=InputFile(media_file),
                        caption="🎥 Here's your video!"
                    )
                else:
                    update.message.reply_photo(
                        photo=InputFile(media_file),
                        caption="📸 Here's your photo!"
                    )
        except Exception as e:
            update.message.reply_text("❌ Failed to send media. Please try again.")
            logger.error(f"Error sending media to {user_id}", exc_info=True)
    else:
        update.message.reply_text("❌ Failed to download media. Please check the URL and try again.")

def inline_query(update: Update, context):
    user_id = update.inline_query.from_user.id
    query = update.inline_query.query.strip()
    
    if not query:
        return
    
    if not query.startswith(('http://', 'https://')):
        update.inline_query.answer([
            InlineQueryResultArticle(
                id="invalid_url",
                title="Invalid URL",
                input_message_content=InputTextMessageContent("Please provide a valid Instagram URL starting with http:// or https://")
            )
        ])
        return
    
    if not check_rate_limit(user_id):
        update.inline_query.answer([
            InlineQueryResultArticle(
                id="rate_limit",
                title="Too many requests",
                input_message_content=InputTextMessageContent("⚠️ Please wait a minute before making more requests")
            )
        ])
        return
    
    clean_old_files()
    
    result = download_instagram_media(query, user_id)
    if not result:
        update.inline_query.answer([
            InlineQueryResultArticle(
                id="error",
                title="Download failed",
                input_message_content=InputTextMessageContent("❌ Failed to download media. Please try a different link.")
            )
        ])
        return
    
    try:
        with open(result['path'], 'rb') as media_file:
            if result['is_video']:
                video_msg = context.bot.send_video(
                    chat_id=CHAT_ID,
                    video=InputFile(media_file),
                    disable_notification=True
                )
                update.inline_query.answer([
                    InlineQueryResultCachedVideo(
                        id="1",
                        title=result['title'],
                        video_file_id=video_msg.video.file_id,
                        caption="🎥 Here's your video!"
                    )
                ])
            else:
                photo_msg = context.bot.send_photo(
                    chat_id=CHAT_ID,
                    photo=InputFile(media_file),
                    disable_notification=True
                )
                update.inline_query.answer([
                    InlineQueryResultCachedPhoto(
                        id="1",
                        title=result['title'],
                        photo_file_id=photo_msg.photo[0].file_id,
                        caption="📸 Here's your photo!"
                    )
                ])
    except Exception as e:
        logger.error("Error processing inline query", exc_info=True)
        update.inline_query.answer([
            InlineQueryResultArticle(
                id="error",
                title="Processing failed",
                input_message_content=InputTextMessageContent("❌ Failed to process media. Please try again.")
            )
        ])

def scheduled_cleanup():
    """Periodically clean up old files"""
    while True:
        time.sleep(CLEANUP_INTERVAL)
        clean_old_files()

def main():
    cleanup_thread = threading.Thread(target=scheduled_cleanup, daemon=True)
    cleanup_thread.start()
    
    updater = Updater(BOT_TOKEN, use_context=True)
    dp = updater.dispatcher
    
    dp.add_handler(CommandHandler("start", start))
    dp.add_handler(MessageHandler(Filters.text & ~Filters.command, handle_message))
    dp.add_handler(InlineQueryHandler(inline_query))
    
    logger.info("Bot is running with multi-user support...")
    updater.start_polling()
    updater.idle()

if __name__ == "__main__":
    main()
