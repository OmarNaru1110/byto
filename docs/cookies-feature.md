# Cookies Feature Guide

This document explains what the Cookies feature is in Byto, when to use it, and how to configure it correctly.

## What is the Cookies feature?

Some websites require a logged-in session to access certain videos, playlists, age-restricted content, memberships, or private media.

The Cookies feature lets Byto pass your authenticated browser session to yt-dlp, so downloads can run with your account access.

In Byto, cookies can be provided in two ways:

1. Upload a cookies.txt file (recommended)
2. Choose a browser and let Byto extract cookies directly

## How to use method 1: Upload cookies.txt (recommended)

1. Export your cookies.txt from your browser using a trusted cookies exporter extension/tool.
2. In Byto, open Add Download.
3. Enable Cookies Options.
4. Select Upload Cookies File (Recommended).
5. Paste the path or click the folder button and pick your cookies.txt file.
6. Add to Queue and start downloads.

Video tutorial for this flow:
https://youtu.be/4Yguy17gQA0

## How to use method 2: Choose Browser

1. Open Add Download.
2. Enable Cookies Options.
3. Select Choose Browser.
4. Pick your browser from the list.
5. Add to Queue and start downloads.

## Important notes and best practices

- Uploading cookies.txt is usually more consistent and reliable than choosing browser directly.
- If you use Choose Browser, fully close that browser before starting downloads. Open browser processes can lock cookie databases and cause extraction failures.
- If authenticated downloads fail, refresh/export a new cookies.txt and try again.
- Keep cookies private. Anyone with your cookies may be able to use your active session.
- Cookies can expire. If downloads suddenly stop working, re-export cookies and update the file in Byto.
- Use cookies only when needed for authenticated or restricted content.

## Quick troubleshooting

- Error or access denied on restricted content:
  - Re-login in browser and export a fresh cookies.txt.
  - Prefer Upload Cookies File over Choose Browser.
- Browser method fails:
  - Close all browser windows and background browser processes, then retry.
- Previously working setup stops working:
  - Your session likely expired. Export new cookies and test again.

## Summary

For the best success rate, use Upload Cookies File with a fresh cookies.txt file, and keep Choose Browser as a secondary option when needed.
