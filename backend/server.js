import express from "express";
import dotenv from "dotenv";
import multer from "multer";
import bodyParser from "body-parser";
import fs from 'fs/promises';
import path from "path";
import { fileURLToPath } from "url";
import mongoose from "mongoose";
import Pothole from "./models/potholes.js";


dotenv.config();
const app = express();


const PORT = process.env.port;
const DB_PASS = process.env.mongoPass;
const DB_STRING = `mongodb+srv://badar:${DB_PASS}@smap-cluster.r26ys.mongodb.net/?retryWrites=true&w=majority&appName=smap-cluster`;


const RADIUS = 2;   
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

mongoose
  .connect(DB_STRING)
  .then(() => {
    console.log("Connected to DB!");
    app.listen(PORT, () => {
      console.log(`Listening on port : ${PORT}`);
    });
  })
  .catch((err) => {
    console.log(`Errored : ${err}`);
  });


app.use(bodyParser.json());
app.use(bodyParser.urlencoded({ extended: true }));


const upload = multer({
  dest: 'images/', // Directory to store uploaded files
  fileFilter: (req, file, cb) => {
    // Accept only JPEG images
    if (file.mimetype === 'image/jpeg') {
      cb(null, true);
    } else {
      cb(new Error('Only JPEG images are allowed!'), false);
    }
  },
});



app.get("/nearby", async (req, res) => {
  const lat = req.query.latitude;
  const long = req.query.longitude;
  console.log(lat, long);

  try {
    const potholes = await Pothole.find({
      location: {
        $nearSphere: {
          $geometry: {
            type: 'Point',
            coordinates: [lat, long] // Search around this point
          },
          $maxDistance: RADIUS // Max distance in meters
        }
      }
    });

    res.send(potholes)
  } catch (err) {
    console.error('Error querying potholes:', err);
  }
});

app.post('/report', async (req, res) =>{
  const lat = req.query.latitude;
  const long = req.query.latitude;

  try {
    // Create a new Pothole instance
    const newPothole = new Pothole({
      location: {
        type: 'Point',             
        coordinates: [lat, long]  
      },
      markedBy: "client", 
    });

    // Save the document to the database
    await newPothole.save();
    console.log('Pothole saved successfully!');
  } catch (err) {
    console.error('Error saving pothole:', err);
  }
})


app.post('/upload', upload.single('image'), async (req, res) => {
  const { latitude, longitude } = req.body;

  if (!req.file) {
    return res.status(400).json({ error: 'No file uploaded' });
  }

  const tempPath = req.file.path;
  const targetPath = path.join(__dirname, 'images', req.file.originalname);

  try {
    // Move the file to the target directory
    await fs.rename(tempPath, targetPath);

    console.log(`Image uploaded: ${req.file.originalname}`);
    console.log(`Latitude: ${latitude}, Longitude: ${longitude}`);

    res.status(200).json({
      message: 'Image uploaded successfully with location data',
      fileName: req.file.originalname,
      latitude,
      longitude,
    });
  } catch (error) {
    console.error('Error saving the file:', error);
    res.status(500).json({ error: 'Error saving the file' });
  }
});

