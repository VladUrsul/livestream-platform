# Set your registry name
$registry = "streamr-registry"

# Step 1: Get all repositories
$repos = doctl registry repository list --registry $registry | Select-Object -Skip 1 | ForEach-Object { ($_ -split "\s+")[0] }

foreach ($repo in $repos) {
    Write-Host "Cleaning repository: $repo"

    # Step 2: Delete all tags
    $tags = doctl registry repository list-tags $repo --registry $registry | Select-Object -Skip 1 | ForEach-Object { ($_ -split "\s+")[0] }
    foreach ($tag in $tags) {
        Write-Host "Deleting tag: ${repo}:${tag}"
        doctl registry repository delete-tag $repo $tag -f
    }

    # Step 3: Delete all remaining manifests
    $manifests = doctl registry repository list-manifests $repo --registry $registry | Select-Object -Skip 1 | ForEach-Object { ($_ -split "\s+")[0] }
    foreach ($digest in $manifests) {
        Write-Host "Deleting manifest: $digest"
        try {
            doctl registry repository delete-manifest $repo $digest -f
        } catch {
            Write-Host "Skipping manifest $digest (probably referenced)"
        }
    }
}

# Step 4: Run garbage collection
Write-Host "Starting garbage collection..."
doctl registry garbage-collection start $registry --include-untagged-manifests --force