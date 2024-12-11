import express from "express";
import dotenv from "dotenv";
import multer from "multer";
import bodyParser from "body-parser";
import fs from 'fs/promises';
import path from "path";
import { fileURLToPath } from "url";
import mongoose from "mongoose";
import Pothole from "./models/potholes.js";
import predict from "./prediction/predict.js";


dotenv.config();
const app = express();


const PORT = process.env.port;
const DB_PASS = process.env.mongoPass;
const DB_STRING = `mongodb+srv://badar:${DB_PASS}@smap-cluster.r26ys.mongodb.net/?retryWrites=true&w=majority&appName=smap-cluster`;


const RADIUS = 1000;   
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

async function recordPothole(coords) {
  try {
    const newPothole = new Pothole({
      location: {
        type: 'Point',             
        coordinates: [coords[0], coords[1]]  
      },
      markedBy: "client", 
    });

    // Save the document to the database
    await newPothole.save();
    console.log('Pothole saved successfully!');
  } catch (err) {
    console.error('Error saving pothole:', err);
  }
}



app.get("/nearby", async (req, res) => {
  const lat = parseFloat(req.query.latitude);
  const long = parseFloat(req.query.longitude);

  if (isNaN(lat) || isNaN(long)) {
    return res.status(400).send({ error: "Invalid latitude or longitude" });
  }

  console.log(`Fetching nearby potholes around lat: ${lat}, long: ${long}`);

  try {
    const potholes = await Pothole.find({
      location: {
        $nearSphere: {
          $geometry: {
            type: 'Point',
            coordinates: [long, lat] 
          },
          $maxDistance: RADIUS // Max distance in meters
        }
      }
    });// Select the relevant fields

    console.log(potholes)

    // Map the potholes to a new structure with latitude and longitude
    const nearbyPotholes = potholes.map(pothole => ({
      latitude: pothole.location.coordinates[1], // Latitude is the second element
      longitude: pothole.location.coordinates[0], // Longitude is the first element
      markedBy: pothole.markedBy,
      reportedAt: pothole.reportedAt
    }));

    // Send the response in the expected format
    res.json({ nearby_potholes: nearbyPotholes });

  } catch (err) {
    console.error('Error querying potholes:', err);
    res.status(500).send({ error: "Internal Server Error" });
  }
});

app.post('/report', async (req, res) =>{
  const lat = req.query.latitude;
  const long = req.query.latitude;

  await recordPothole([lat, long]);
})


app.post('/upload', upload.single('image'), async (req, res) => {
  const { latitude, longitude } = req.body;

  if (!req.file) {
    return res.status(400).json({ error: 'No file uploaded' });
  }

  const tempPath = req.file.path;
  const targetPath = path.join(__dirname, 'images', req.file.originalname);
  const potholePath = path.join(__dirname, "images/potholes", req.file.originalname);
  const npotholePath = path.join(__dirname, "images/npotholes", req.file.originalname);

  try {
    // Move the file to the target directory
    await fs.rename(tempPath, targetPath);
    const result = parseInt(await predict(targetPath), 10);
    console.log(result)

    switch (result) {
      case 0:
        await fs.rename(targetPath, npotholePath);
        break;
      case 1:
        await fs.rename(targetPath, potholePath);
        await recordPothole([latitude, longitude])
        break;
    }

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

