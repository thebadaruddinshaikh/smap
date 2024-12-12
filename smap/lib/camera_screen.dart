import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;
import 'dart:convert'; // For JSON decoding
import 'package:http_parser/http_parser.dart';
import 'package:geolocator/geolocator.dart';
import 'package:google_maps_flutter/google_maps_flutter.dart';
import 'package:flutter_compass/flutter_compass.dart';
import 'package:camera/camera.dart';
import 'dart:async';

class CameraScreen extends StatefulWidget {
  final CameraDescription camera;

  const CameraScreen({Key? key, required this.camera}) : super(key: key);

  @override
  _CameraScreenState createState() => _CameraScreenState();
}

class _CameraScreenState extends State<CameraScreen> {
  late CameraController _controller;
  late Future<void> _initializeControllerFuture;
  Timer? _timer;

  final int intervalSeconds = 500;
  late GoogleMapController mapController;
  LatLng? _currentPosition; // Make this nullable to indicate no location yet
  double _heading = 0.0; // Initial heading
  late double _speed;

  // BitmapDescriptor? _locationIcon; // Custom location icon
  Set<Marker> _potholeMarkers = {}; // Markers for potholes

  @override
  void initState() {
    super.initState();
    _controller = CameraController(
      widget.camera,
      ResolutionPreset.low,
    );
    _initializeControllerFuture = _controller.initialize();
    _controller.setFlashMode(FlashMode.off);

    // Load custom icon for the user's location
    // await _loadLocationIcon();

    // Start the timer for automatic photo capture
    _timer = Timer.periodic(Duration(milliseconds: intervalSeconds), (timer) {
      _takePictureAndSend();
    });

    _getLocationUpdates(); // Start location updates and wait for the first location
    _getHeadingUpdates(); // Start listening to heading updates
  }

  @override
  void dispose() {
    _controller.dispose();
    _timer?.cancel();
    super.dispose();
  }

  // Future<void> _loadLocationIcon() async {
  //   _locationIcon = await BitmapDescriptor.fromAssetImage(
  //     ImageConfiguration(size: Size(48, 48)), // Customize size as needed
  //     'assets/location_icon.png', // Replace with your icon asset
  //   );
  // }

  Future<void> _takePictureAndSend() async {
    if (_currentPosition == null || _speed < 5)
      return; // Wait until location is available

    try {
      await _initializeControllerFuture;

      // Take the picture and get the file path
      final image = await _controller.takePicture();

      // Send the image and location data to the server
      await _sendImageToApi(
          image.path, _currentPosition!.latitude, _currentPosition!.longitude);
    } catch (e) {
      print("Error taking picture: $e");
    }
  }

  Future<Position> _getLocation() async {
    return await Geolocator.getCurrentPosition();
  }

  void _getLocationUpdates() async {
    try {
      Position position = await _getLocation();
      setState(() {
        _currentPosition = LatLng(position.latitude, position.longitude);
      });

      // Fetch potholes nearby
      _fetchNearbyPotholes();

      Geolocator.getPositionStream().listen((Position position) {
        setState(() {
          _currentPosition = LatLng(position.latitude, position.longitude);
          _speed = position.speed;
        });
        print(
          "This is the speed ::  $_speed",
        );
        mapController.animateCamera(
          CameraUpdate.newLatLng(_currentPosition!),
        );

        // Fetch nearby potholes again based on the updated location
        _fetchNearbyPotholes();
      });
    } catch (e) {
      print("Error getting location: $e");
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text("Failed to get location: $e")),
      );
    }
  }

  void _getHeadingUpdates() {
    FlutterCompass.events!.listen((CompassEvent event) {
      setState(() {
        _heading = event.heading ?? 0; // Get the heading or default to 0
      });
    });
  }

  Future<void> _fetchNearbyPotholes() async {
    if (_currentPosition == null) return;

    final url = Uri.parse("http://98.11.205.187:45670/nearby");

    try {
      final response = await http.get(url.replace(queryParameters: {
        'latitude': _currentPosition!.latitude.toString(),
        'longitude': _currentPosition!.longitude.toString(),
      }));

      if (response.statusCode == 200) {
        // Decode the JSON response
        final Map<String, dynamic> data = json.decode(response.body);

        // Ensure the response has a 'potholes' key that is a list
        if (data.containsKey('nearby_potholes') &&
            data['nearby_potholes'] is List) {
          List<dynamic> potholes = data['nearby_potholes'];

          Set<Marker> markers = potholes.map<Marker>((item) {
            final lat = item['latitude'];
            final lng = item['longitude'];
            return Marker(
              markerId: MarkerId("pothole_${lat}_$lng"),
              position: LatLng(lat, lng),
              infoWindow: InfoWindow(title: "Pothole"),
              icon: BitmapDescriptor.defaultMarkerWithHue(
                  BitmapDescriptor.hueRed),
            );
          }).toSet();

          setState(() {
            _potholeMarkers = markers;
          });
        } else {
          print(
              "Invalid data format: 'potholes' key is missing or not a list.");
        }
      } else {
        print("Failed to load nearby potholes: ${response.statusCode}");
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
              content: Text(
                  "Failed to load nearby potholes: ${response.statusCode}")),
        );
      }
    } catch (e) {
      print("Error fetching nearby potholes: $e");
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text("Error fetching nearby potholes: $e")),
      );
    }
  }

  Future<void> _sendImageToApi(
      String imagePath, double latitude, double longitude) async {
    final uri = Uri.parse("http://98.11.205.187:45670/upload");
    final request = http.MultipartRequest("POST", uri);
    request.files.add(await http.MultipartFile.fromPath(
      'image',
      imagePath,
      contentType: MediaType('image', 'jpeg'),
    ));

    request.fields['latitude'] = latitude.toString();
    request.fields['longitude'] = longitude.toString();

    try {
      final response = await request.send();
      if (response.statusCode == 200) {
        print("Image uploaded successfully with location data");
      } else {
        print("Failed to upload image: ${response.statusCode}");
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
              content: Text("Failed to upload image: ${response.statusCode}")),
        );
      }
    } catch (e) {
      print("Error uploading image: $e");
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text("Error uploading image: $e")),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Center(child: Text("Smap"))),
      body: _currentPosition == null
          ? const Center(
              child:
                  CircularProgressIndicator()) // Show loading until location is available
          : GoogleMap(
              onMapCreated: (GoogleMapController controller) {
                mapController = controller;
              },
              initialCameraPosition: CameraPosition(
                target: _currentPosition!,
                zoom: 14.0,
              ),
              markers: {
                Marker(
                  markerId: MarkerId("currentLocation"),
                  position: _currentPosition!,
                  icon: BitmapDescriptor.defaultMarker,
                  rotation: _heading, // Rotate marker based on heading
                  anchor: Offset(0.5, 0.5), // Center the icon
                ),
                ..._potholeMarkers, // Add pothole markers
              },
            ),
    );
  }
}
