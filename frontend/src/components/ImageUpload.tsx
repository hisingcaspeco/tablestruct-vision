import React, { useState } from "react";
import { Button, Group, Loader, Text, Textarea, Image, Center } from "@mantine/core";
import { Dropzone } from "@mantine/dropzone";
import {IconPhone, IconPhoto, IconUpload, IconX} from "@tabler/icons-react";

export const ImageUpload: React.FC = () => {
    const [selectedFile, setSelectedFile] = useState<File | null>(null);
    const [previewURL, setPreviewURL] = useState<string | null>(null);
    const [responseJSON, setResponseJSON] = useState<string | null>(null);
    const [isLoading, setIsLoading] = useState<boolean>(false);
    const [error, setError] = useState<string | null>(null);

    // Handle file selection & preview
    const handleFileChange = (files: File[]) => {
        if (files.length > 0) {
            const file = files[0];
            setSelectedFile(file);
            setPreviewURL(URL.createObjectURL(file)); // Show preview
            setResponseJSON(null); // Reset previous response
            setError(null);
        }
    };

    // Handle file upload
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

            // ✅ Parse JSON inside the response field if necessary
            if (data.response) {
                try {
                    const parsedJson = JSON.parse(data.response);
                    setResponseJSON(JSON.stringify(parsedJson, null, 2));
                } catch (error) {
                    console.error("Failed to parse JSON:", error);
                    setResponseJSON("Invalid JSON format received.");
                }
            } else {
                setResponseJSON("No response from server.");
            }
        } catch (err) {
            setError((err as Error).message);
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div style={{ maxWidth: 600, margin: "auto" }}>
            <Text size="xl" w={500} align="center">
                Upload a Restaurant Layout
            </Text>

            {/* Dropzone for File Upload */}
            <Dropzone
                onDrop={handleFileChange}
                accept={["image/png", "image/jpeg", "image/jpg", "image/gif", "image/svg+xml", "image/webp"]}
                multiple={false}
                maxSize={5 * 1024 * 1024} // 5MB max
                style={{
                    marginTop: 20,
                    padding: 20,
                    border: "2px dashed #1e88e5",
                    borderRadius: 10,
                    textAlign: "center",
                }}
            >
                <Group justify="center" gap="xl" mih={220} style={{ pointerEvents: 'none' }}>
                    <Dropzone.Accept>
                        <IconUpload size={52} color="var(--mantine-color-blue-6)" stroke={1.5} />
                    </Dropzone.Accept>
                    <Dropzone.Reject>
                        <IconX size={52} color="var(--mantine-color-red-6)" stroke={1.5} />
                    </Dropzone.Reject>
                    <Dropzone.Idle>
                        <IconPhoto size={52} color="var(--mantine-color-dimmed)" stroke={1.5} />
                    </Dropzone.Idle>

                    <div>
                        <Text size="xl" inline>
                            Drag images here or click to select files
                        </Text>
                        <Text size="sm" c="dimmed" inline mt={7}>
                            Attach as many files as you like, each file should not exceed 5mb
                        </Text>
                    </div>
                </Group>
            </Dropzone>

            {/* Image Preview */}
            {previewURL && (
                <Center mt="md">
                    <Image src={previewURL} alt="Preview" radius="md" width={300} />
                </Center>
            )}

            {/* Upload Button */}
            <Group position="center" mt="md">
                <Button onClick={handleUpload} disabled={isLoading}>
                    {isLoading ? <Loader size="sm" /> : "Upload & Analyze"}
                </Button>
            </Group>

            {/* Loading Text */}
            {isLoading && (
                <Text align="center" color="blue" mt="md">
                    Processing the image... Please wait.
                </Text>
            )}

            {/* Error Message */}
            {error && (
                <Text align="center" color="red" mt="md">
                    {error}
                </Text>
            )}

            {/* JSON Output */}
            {responseJSON && (
                <Textarea
                    label="Detected Layout (JSON)"
                    value={responseJSON}
                    readOnly
                    autosize
                    minRows={6}
                    style={{ marginTop: 20 }}
                />
            )}
        </div>
    );
};
