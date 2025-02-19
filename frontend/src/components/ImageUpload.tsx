import React, { useState } from "react";

export const ImageUpload: React.FC = () => {
    const [selectedFile, setSelectedFile] = useState<File | null>(null);
    const [previewURL, setPreviewURL] = useState<string | null>(null);
    const [responseJSON, setResponseJSON] = useState<string | null>(null);
    const [isLoading, setIsLoading] = useState<boolean>(false);
    const [error, setError] = useState<string | null>(null);

    // Handle file selection & preview
    const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        if (event.target.files && event.target.files.length > 0) {
            const file = event.target.files[0];
            setSelectedFile(file);
            setPreviewURL(URL.createObjectURL(file)); // Show preview
            setResponseJSON(null); // Reset previous response
            setError(null);
        }
    };

    // Handle form submission
    const handleUpload = async () => {
        if (!selectedFile) {
            setError("Please select an image first.");
            return;
        }

        setIsLoading(true);
        setError(null);
        setResponseJSON(null);

        const formData = new FormData();
        formData.append("image", selectedFile);

        try {
            const res = await fetch("http://localhost:8080/upload", {
                method: "POST",
                body: formData,
            });

            if (!res.ok) {
                throw new Error(`Upload failed: ${res.statusText}`);
            }

            const data = await res.json();
            setResponseJSON(JSON.stringify(data, null, 2)); // Pretty-print JSON
        } catch (err) {
            setError((err as Error).message);
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="container">
            <h2>Upload a Restaurant Layout</h2>

            {/* File Input */}
            <input type="file" accept="image/*" onChange={handleFileChange} />

            {/* Image Preview */}
            {previewURL && (
                <div className="preview-container">
                    <h3>Preview:</h3>
                    <img src={previewURL} alt="Preview" className="preview-image" />
                </div>
            )}

            {/* Upload Button */}
            <button onClick={handleUpload} disabled={isLoading}>
                {isLoading ? "Generating..." : "Upload & Analyze"}
            </button>

            {/* Loading Text */}
            {isLoading && <p>Processing the image... Please wait.</p>}

            {/* Error Message */}
            {error && <p className="error">{error}</p>}

            {/* JSON Output */}
            {responseJSON && (
                <div>
                    <h3>Detected Layout (JSON):</h3>
                    <textarea readOnly value={responseJSON} className="json-output" />
                </div>
            )}
        </div>
    );
};
