package mongodbatlas

import (
	"context"
	"fmt"

	atlasSDK "go.mongodb.org/atlas-sdk/v20231001001/admin"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// This datasource does not have a resource: we tested it manually
func dataSourceMongoDBAtlasCloudSharedTierRestoreJobs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMongoDBAtlasCloudSharedTierRestoreJobRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"cluster_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"results": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"job_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"target_project_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"target_deployment_item_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"snapshot_url": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"snapshot_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"snapshot_finished_date": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"restore_scheduled_date": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"restore_finished_date": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"delivery_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"expiration_date": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"total_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceMongoDBAtlasCloudSharedTierRestoreJobRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*MongoDBClient).AtlasV2

	projectID := d.Get("project_id").(string)
	clusterName := d.Get("cluster_name").(string)

	jobs, _, err := conn.SharedTierRestoreJobsApi.ListSharedClusterBackupRestoreJobs(ctx, clusterName, projectID).Execute()
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting shared tier restore jobs for cluster '%s': %w", clusterName, err))
	}

	if err := d.Set("results", flattenShardTierRestoreJobs(jobs.Results)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `results`: %w", err))
	}

	if err := d.Set("total_count", jobs.TotalCount); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `total_count`: %w", err))
	}

	d.SetId(id.UniqueId())
	return nil
}

func flattenShardTierRestoreJobs(sharedTierJobs []atlasSDK.TenantRestore) []map[string]any {
	if len(sharedTierJobs) == 0 {
		return nil
	}

	results := make([]map[string]any, len(sharedTierJobs))
	for i := range sharedTierJobs {
		sharedTierJob := &sharedTierJobs[i]
		results[i] = map[string]any{
			"job_id":                      sharedTierJob.Id,
			"status":                      sharedTierJob.Status,
			"target_project_id":           sharedTierJob.TargetProjectId,
			"target_deployment_item_name": sharedTierJob.TargetDeploymentItemName,
			"snapshot_url":                sharedTierJob.SnapshotUrl,
			"snapshot_id":                 sharedTierJob.SnapshotId,
			"delivery_type":               sharedTierJob.DeliveryType,
			"snapshot_finished_date":      sharedTierJob.GetSnapshotFinishedDate().String(),
			"restore_scheduled_date":      sharedTierJob.GetRestoreScheduledDate().String(),
			"restore_finished_date":       sharedTierJob.GetRestoreFinishedDate().String(),
			"expiration_date":             sharedTierJob.GetExpirationDate().String(),
		}
	}

	return results
}
