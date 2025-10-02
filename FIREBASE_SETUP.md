# Firebase Leaderboard Setup

This document explains how to set up Firebase Firestore for the sailing game leaderboard.

## Prerequisites

- Google account
- Access to Firebase console (https://console.firebase.google.com/)

## Firebase Project Setup

1. **Create a new Firebase project**:
   - Go to https://console.firebase.google.com/
   - Click "Add project"
   - Enter project name (e.g., "gosailing-leaderboard")
   - Continue through setup (disable Google Analytics if not needed)

2. **Enable Firestore Database**:
   - In Firebase console, go to "Firestore Database"
   - Click "Create database"
   - Choose "Start in test mode" (allows read/write for 30 days)
   - Select a location close to your users

3. **Get Firebase configuration**:
   - Go to Project Settings (gear icon)
   - Scroll down to "Your apps"
   - Click "Add app" > Web app (</> icon)
   - Register app with nickname (e.g., "sailing-game")
   - Copy the `firebaseConfig` object

4. **Update web/index.html**:
   Replace the placeholder config in `web/index.html`:
   ```javascript
   const firebaseConfig = {
       apiKey: "your-actual-api-key",
       authDomain: "your-project.firebaseapp.com",
       projectId: "your-actual-project-id",
       storageBucket: "your-project.appspot.com",
       messagingSenderId: "123456789",
       appId: "your-actual-app-id"
   };
   ```

## Firestore Security Rules

For production, update Firestore rules to be more restrictive:

```javascript
rules_version = '2';
service cloud.firestore {
  match /databases/{database}/documents {
    // Allow read access to race_results
    match /race_results/{document} {
      allow read: if true;

      // Allow write only if data is valid
      allow create: if request.auth == null &&
                   isValidRaceResult(request.resource.data);
    }
  }
}

function isValidRaceResult(data) {
  return data.keys().hasAll(['player_name', 'race_time_seconds', 'seconds_late', 'speed_percentage', 'mark_rounded', 'timestamp']) &&
         data.player_name is string &&
         data.player_name.size() > 0 &&
         data.player_name.size() <= 50 &&
         data.race_time_seconds is number &&
         data.race_time_seconds > 0 &&
         data.race_time_seconds < 3600 && // Max 1 hour
         data.seconds_late is number &&
         data.speed_percentage is number &&
         data.mark_rounded is bool &&
         data.timestamp is number;
}
```

## Database Structure

The leaderboard uses a single Firestore collection:

### Collection: `race_results`
```json
{
  "player_name": "string (max 50 chars)",
  "race_time_seconds": "number (race completion time)",
  "seconds_late": "number (how late at start)",
  "speed_percentage": "number (% of target speed)",
  "mark_rounded": "boolean (completed the course)",
  "timestamp": "number (unix timestamp)"
}
```

## Testing

1. **Test Mode**: Initially set Firestore to test mode for easy development
2. **Local Testing**: The game will work locally; scoreboard just won't persist
3. **WASM Testing**: Deploy to web server and test score submission

## Cost Considerations

Firebase Firestore free tier includes:
- 50K reads per day
- 20K writes per day
- 1GB storage

For a sailing game leaderboard, this should be sufficient for moderate usage.

## Deployment Notes

- The game will work without Firebase (shows local results only)
- Firebase is only active in WASM builds
- Standalone builds ignore leaderboard functionality
- Consider adding rate limiting for production use

## Troubleshooting

- **"Firebase not initialized"**: Check console for JavaScript errors
- **Permission denied**: Update Firestore security rules
- **Network errors**: Check Firebase project settings and API keys
- **CORS issues**: Ensure proper Firebase domain configuration
