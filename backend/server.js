import express from "express";
import dotenv from "dotenv";

import authRouter from "./authRouter.js";

dotenv.config();
const app = express();

const PORT = process.env.PORT;

app.use("/auth", authRouter);

app.get("/nearby", (req, res) => {});
app.post("/upload", (req, res) => {});

app.listen(PORT, () => {
  console.log(`Listening on Port : ${PORT}`);
});
