import mongoose from "mongoose";

const potholesSchema = mongoose.Schema({
    location : {
        type: {
            type: String,
            enum : ['Point'],
            requried : true
        },

        coordinates : {
            type: [Number],
            required : true
        },
    },
    markedBy: {
        type: String,
        requred : true
    },
    reportedAt : {
        type: Date,
        default: Date.now
    }
})

potholesSchema.index({location: "2dsphere"})
const Pothole = mongoose.model("pothole", potholesSchema)

export default Pothole;