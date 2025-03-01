package mongodbatlas

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spf13/cast"
	matlas "go.mongodb.org/atlas/mongodbatlas"
)

func resourceMongoDBAtlasCloudBackupSnapshot() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMongoDBAtlasCloudBackupSnapshotCreate,
		ReadContext:   resourceMongoDBAtlasCloudBackupSnapshotRead,
		DeleteContext: resourceMongoDBAtlasCloudBackupSnapshotDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceMongoDBAtlasCloudBackupSnapshotImportState,
		},
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cluster_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"retention_in_days": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"expires_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"master_key_uuid": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"mongod_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"snapshot_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"storage_size_bytes": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"snapshot_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cloud_provider": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"members": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cloud_provider": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"replica_set_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"replica_set_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"snapshot_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceMongoDBAtlasCloudBackupSnapshotRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Get client connection.
	conn := meta.(*MongoDBClient).Atlas
	ids := decodeStateID(d.Id())

	requestParameters := &matlas.SnapshotReqPathParameters{
		SnapshotID:  ids["snapshot_id"],
		GroupID:     ids["project_id"],
		ClusterName: ids["cluster_name"],
	}

	snapshot, resp, err := conn.CloudProviderSnapshots.GetOneCloudProviderSnapshot(context.Background(), requestParameters)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}

		return diag.FromErr(fmt.Errorf("error getting snapshot Information: %s", err))
	}

	if err = d.Set("snapshot_id", snapshot.ID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `snapshot_id` for snapshot (%s): %s", ids["snapshot_id"], err))
	}

	if err = d.Set("created_at", snapshot.CreatedAt); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `created_at` for snapshot (%s): %s", ids["snapshot_id"], err))
	}

	if err = d.Set("expires_at", snapshot.ExpiresAt); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `expires_at` for snapshot (%s): %s", ids["snapshot_id"], err))
	}

	if err = d.Set("master_key_uuid", snapshot.MasterKeyUUID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `master_key_uuid` for snapshot (%s): %s", ids["snapshot_id"], err))
	}

	if err = d.Set("mongod_version", snapshot.MongodVersion); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `mongod_version` for snapshot (%s): %s", ids["snapshot_id"], err))
	}

	if err = d.Set("snapshot_type", snapshot.SnapshotType); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `snapshot_type` for snapshot (%s): %s", ids["snapshot_id"], err))
	}

	if err = d.Set("status", snapshot.Status); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `status` for snapshot (%s): %s", ids["snapshot_id"], err))
	}

	if err = d.Set("storage_size_bytes", snapshot.StorageSizeBytes); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `storage_size_bytes` for snapshot (%s): %s", ids["snapshot_id"], err))
	}

	if err = d.Set("type", snapshot.Type); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `type` for snapshot (%s): %s", ids["snapshot_id"], err))
	}

	if err = d.Set("cloud_provider", snapshot.CloudProvider); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `cloud_provider` for snapshot (%s): %s", ids["snapshot_id"], err))
	}

	if err = d.Set("members", flattenCloudMembers(snapshot.Members)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `members` for snapshot (%s): %s", ids["snapshot_id"], err))
	}

	if err = d.Set("replica_set_name", snapshot.ReplicaSetName); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `replica_set_name` for snapshot (%s): %s", ids["snapshot_id"], err))
	}

	if err = d.Set("snapshot_ids", snapshot.SnapshotsIds); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `snapshot_ids` for snapshot (%s): %s", ids["snapshot_id"], err))
	}

	return nil
}

func resourceMongoDBAtlasCloudBackupSnapshotCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Get client connection.
	conn := meta.(*MongoDBClient).Atlas

	requestParameters := &matlas.SnapshotReqPathParameters{
		GroupID:     d.Get("project_id").(string),
		ClusterName: d.Get("cluster_name").(string),
	}

	snapshotReq := &matlas.CloudProviderSnapshot{
		Description:     d.Get("description").(string),
		RetentionInDays: d.Get("retention_in_days").(int),
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"CREATING", "UPDATING", "REPAIRING", "REPEATING"},
		Target:     []string{"IDLE"},
		Refresh:    resourceClusterRefreshFunc(ctx, d.Get("cluster_name").(string), d.Get("project_id").(string), conn),
		Timeout:    10 * time.Minute,
		MinTimeout: 10 * time.Second,
		Delay:      3 * time.Minute,
	}

	// Wait, catching any errors
	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	snapshot, _, err := conn.CloudProviderSnapshots.Create(ctx, requestParameters, snapshotReq)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error taking a snapshot: %s", err))
	}

	requestParameters.SnapshotID = snapshot.ID

	stateConf = &retry.StateChangeConf{
		Pending:    []string{"queued", "inProgress"},
		Target:     []string{"completed", "failed"},
		Refresh:    resourceCloudBackupSnapshotRefreshFunc(ctx, requestParameters, conn),
		Timeout:    1 * time.Hour,
		MinTimeout: 60 * time.Second,
		Delay:      1 * time.Minute,
	}

	// Wait, catching any errors
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(encodeStateID(map[string]string{
		"project_id":   d.Get("project_id").(string),
		"cluster_name": d.Get("cluster_name").(string),
		"snapshot_id":  snapshot.ID,
	}))

	return resourceMongoDBAtlasCloudBackupSnapshotRead(ctx, d, meta)
}

func resourceMongoDBAtlasCloudBackupSnapshotDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Get client connection.
	conn := meta.(*MongoDBClient).Atlas
	ids := decodeStateID(d.Id())

	requestParameters := &matlas.SnapshotReqPathParameters{
		SnapshotID:  ids["snapshot_id"],
		GroupID:     ids["project_id"],
		ClusterName: ids["cluster_name"],
	}

	_, err := conn.CloudProviderSnapshots.Delete(ctx, requestParameters)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting a snapshot (%s): %s", ids["snapshot_id"], err))
	}

	return nil
}

func resourceCloudBackupSnapshotRefreshFunc(ctx context.Context, requestParameters *matlas.SnapshotReqPathParameters, client *matlas.Client) retry.StateRefreshFunc {
	return func() (any, string, error) {
		c, resp, err := client.CloudProviderSnapshots.GetOneCloudProviderSnapshot(ctx, requestParameters)

		switch {
		case err != nil:
			return nil, "failed", err
		case resp.StatusCode == http.StatusNotFound:
			return "", "DELETED", nil
		case c.Status == "failed":
			return nil, c.Status, fmt.Errorf("error creating MongoDB snapshot(%s) status was: %s", requestParameters.SnapshotID, c.Status)
		}

		if c.Status != "" {
			log.Printf("[DEBUG] status for MongoDB snapshot: %s: %s", requestParameters.SnapshotID, c.Status)
		}

		return c, c.Status, nil
	}
}

func resourceMongoDBAtlasCloudBackupSnapshotImportState(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	conn := meta.(*MongoDBClient).Atlas

	requestParameters, err := splitSnapshotImportID(d.Id())
	if err != nil {
		return nil, err
	}

	u, _, err := conn.CloudProviderSnapshots.GetOneCloudProviderSnapshot(ctx, requestParameters)
	if err != nil {
		return nil, fmt.Errorf("couldn't import snapshot %s in project %s, error: %s", requestParameters.ClusterName, requestParameters.GroupID, err)
	}

	d.SetId(encodeStateID(map[string]string{
		"project_id":   requestParameters.GroupID,
		"cluster_name": requestParameters.ClusterName,
		"snapshot_id":  requestParameters.SnapshotID,
	}))

	if err := d.Set("project_id", requestParameters.GroupID); err != nil {
		log.Printf("[WARN] Error setting project_id for (%s): %s", requestParameters.SnapshotID, err)
	}

	if err := d.Set("cluster_name", requestParameters.ClusterName); err != nil {
		log.Printf("[WARN] Error setting cluster_name for (%s): %s", requestParameters.SnapshotID, err)
	}

	if err := d.Set("description", u.Description); err != nil {
		log.Printf("[WARN] Error setting description for (%s): %s", requestParameters.SnapshotID, err)
	}

	return []*schema.ResourceData{d}, nil
}

func splitSnapshotImportID(id string) (*matlas.SnapshotReqPathParameters, error) {
	var re = regexp.MustCompile(`(?s)^([0-9a-fA-F]{24})-(.*)-([0-9a-fA-F]{24})$`)

	parts := re.FindStringSubmatch(id)

	if len(parts) != 4 {
		return nil, errors.New("import format error: to import a snapshot, use the format {project_id}-{cluster_name}-{snapshot_id}")
	}

	return &matlas.SnapshotReqPathParameters{
		GroupID:     parts[1],
		ClusterName: parts[2],
		SnapshotID:  parts[3],
	}, nil
}

func flattenCloudMember(apiObject *matlas.Member) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	tfMap["cloud_provider"] = apiObject.CloudProvider
	tfMap["id"] = apiObject.ID
	tfMap["replica_set_name"] = apiObject.ReplicaSetName

	return tfMap
}

func flattenCloudMembers(apiObjects []*matlas.Member) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenCloudMember(apiObject))
	}

	return tfList
}

// Support functions moved from deprecated cloud_provider resources

const (
	errorSnapshotBackupPolicyRead    = "error getting a Cloud Provider Snapshot Backup Policy for the cluster(%s): %s"
	errorSnapshotBackupPolicySetting = "error setting `%s` for Cloud Provider Snapshot Backup Policy(%s): %s"
)

func flattenPolicies(policies []matlas.Policy) []map[string]any {
	actionList := make([]map[string]any, 0)
	for _, v := range policies {
		actionList = append(actionList, map[string]any{
			"id":          v.ID,
			"policy_item": flattenPolicyItems(v.PolicyItems),
		})
	}

	return actionList
}

func flattenPolicyItems(items []matlas.PolicyItem) []map[string]any {
	policyItems := make([]map[string]any, 0)
	for _, v := range items {
		policyItems = append(policyItems, map[string]any{
			"id":                 v.ID,
			"frequency_interval": v.FrequencyInterval,
			"frequency_type":     v.FrequencyType,
			"retention_unit":     v.RetentionUnit,
			"retention_value":    v.RetentionValue,
		})
	}

	return policyItems
}

func flattenCloudProviderSnapshotBackupPolicy(ctx context.Context, d *schema.ResourceData, conn *matlas.Client, projectID, clusterName string) ([]map[string]any, error) {
	backupPolicy, res, err := conn.CloudProviderSnapshotBackupPolicies.Get(ctx, projectID, clusterName)
	if err != nil {
		if res.StatusCode == http.StatusNotFound ||
			strings.Contains(err.Error(), "BACKUP_CONFIG_NOT_FOUND") ||
			strings.Contains(err.Error(), "Not Found") ||
			strings.Contains(err.Error(), "404") {
			return []map[string]any{}, nil
		}

		return []map[string]any{}, fmt.Errorf(errorSnapshotBackupPolicyRead, clusterName, err)
	}

	return []map[string]any{
		{
			"cluster_id":               backupPolicy.ClusterID,
			"cluster_name":             backupPolicy.ClusterName,
			"next_snapshot":            backupPolicy.NextSnapshot,
			"reference_hour_of_day":    backupPolicy.ReferenceHourOfDay,
			"reference_minute_of_hour": backupPolicy.ReferenceMinuteOfHour,
			"restore_window_days":      backupPolicy.RestoreWindowDays,
			"update_snapshots":         cast.ToBool(backupPolicy.UpdateSnapshots),
			"policies":                 flattenPolicies(backupPolicy.Policies),
		},
	}, nil
}

func computedCloudProviderSnapshotBackupPolicySchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"cluster_id": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"cluster_name": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"next_snapshot": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"reference_hour_of_day": {
					Type:     schema.TypeInt,
					Computed: true,
				},
				"reference_minute_of_hour": {
					Type:     schema.TypeInt,
					Computed: true,
				},
				"restore_window_days": {
					Type:     schema.TypeInt,
					Computed: true,
				},
				"update_snapshots": {
					Type:     schema.TypeBool,
					Computed: true,
				},
				"policies": {
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"id": {
								Type:     schema.TypeString,
								Computed: true,
							},
							"policy_item": {
								Type:     schema.TypeList,
								Computed: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"id": {
											Type:     schema.TypeString,
											Computed: true,
										},
										"frequency_interval": {
											Type:     schema.TypeInt,
											Computed: true,
										},
										"frequency_type": {
											Type:     schema.TypeString,
											Computed: true,
										},
										"retention_unit": {
											Type:     schema.TypeString,
											Computed: true,
										},
										"retention_value": {
											Type:     schema.TypeInt,
											Computed: true,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
