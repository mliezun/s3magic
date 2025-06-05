package cmd

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/spf13/cobra"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [file_path]",
	Short: "Deletes S3 objects listed in a file.",
	Long: `Deletes multiple S3 objects based on a list provided in a specified file.
Each line in the file should contain the full S3 path for an object to be deleted (e.g., my-bucket/my-object.txt).

Example:
s3magic delete /path/to/objects_to_delete.txt`,
	Args: cobra.ExactArgs(1), // Expect exactly one argument: the file path
	Run: func(cmd *cobra.Command, args []string) {
		// Using the SDK's default configuration, load additional config
		// and credentials values from the environment variables, shared
		// credentials, and shared configuration files
		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			os.Exit(1) // Exit with error status
			return
		}
		// Create an Amazon S3 service client
		client := s3.NewFromConfig(cfg)

		var continuationToken *string
		var allObjects [][]types.Object
		bucket := args[0]

		for {
			output, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
				Bucket:            aws.String(bucket),
				ContinuationToken: continuationToken,
			})
			if err != nil {
				panic(err)
			}

			// Process objects in current page
			allObjects = append(allObjects, output.Contents)
			go deleteObjects(bucket, output.Contents)

			// Check if there are more pages
			if !*output.IsTruncated {
				break
			}

			// Set continuation token for next page
			continuationToken = output.NextContinuationToken
		}

		wg := sync.WaitGroup{}

		for _, objects := range allObjects {
			wg.Add(1)
			go func(objects []types.Object) {
				deleteObjects(bucket, objects)
				wg.Done()
			}(objects)
		}

		wg.Wait()

		// // In a real implementation, you would add logic here to:
		// // 1. Read and parse the specified file.
		// // 2. For each object in the file, interact with the AWS SDK to perform the deletion.
		// // 3. Handle errors (e.g., file not found, malformed file, S3 deletion errors).
		// // 4. Provide progress and summary feedback.

		// // Dummy implementation:
		// // Check if file exists as a basic validation
		// if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// 	fmt.Fprintf(os.Stderr, "Error: File not found at %s\n", filePath)
		// 	os.Exit(1) // Exit with error status
		// 	return
		// }

		// fmt.Printf("Simulating batch deletion of objects listed in: %s\n", filePath)
		// fmt.Println("This is a dummy implementation. No actual deletion will occur.")
	},
}

// deleteObjects deletes all objects from the provided Contents slice
func deleteObjects(bucketName string, objects []types.Object) error {
	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err)
	}

	// Create an Amazon S3 service client
	client := s3.NewFromConfig(cfg)

	// If no objects to delete, return early
	if len(objects) == 0 {
		fmt.Println("No objects to delete")
		return nil
	}

	if len(objects) > 1000 {
		return fmt.Errorf("too many objects")
	}

	// Build the delete request
	var objectsToDelete []types.ObjectIdentifier
	for _, obj := range objects {
		objectsToDelete = append(objectsToDelete, types.ObjectIdentifier{
			Key: obj.Key,
		})
	}

	// Perform the batch delete
	deleteInput := &s3.DeleteObjectsInput{
		Bucket: aws.String(bucketName),
		Delete: &types.Delete{
			Objects: objectsToDelete,
			Quiet:   aws.Bool(false), // Set to true to suppress successful deletion output
		},
	}

	output, err := client.DeleteObjects(context.TODO(), deleteInput)
	if err != nil {
		return fmt.Errorf("failed to delete objects: %w", err)
	}

	// Report results
	fmt.Printf("Successfully deleted %d objects\n", len(output.Deleted))

	// Report any errors
	if len(output.Errors) > 0 {
		fmt.Printf("Failed to delete %d objects:\n", len(output.Errors))
		for _, deleteError := range output.Errors {
			fmt.Printf("  - %s: %s (%s)\n",
				aws.ToString(deleteError.Key),
				aws.ToString(deleteError.Message),
				aws.ToString(deleteError.Code))
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
