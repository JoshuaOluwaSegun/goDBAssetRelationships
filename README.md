# Database Asset Relationship Import Go - [GO](https://golang.org/) Asset Relationship to Hornbill Import Script

## Installation

- Download the archive containing the import executable relevant for your operating system and architecture
- Extract zip into a folder you would like the application to run from e.g. `C:\assetrelationshipimport\`
- Open '''conf.json''' and add in the necessary configration
- Open a Command Line or Terminal session as Administrator
- Change Directory to the folder containing the tool `C:\assetrelationshipimport\`
- Run the command :
  - For Windows Systems: goDBAssetRelationships.exe
  - For *nix Systems: ./goDBAssetRelationships

## Configuration

Example JSON File:

```json
{
    "APIKey": "",
    "InstanceId": "",
    "DBConf": {
        "Driver": "mysql",
        "Server": "127.0.0.1",
        "Database": "assetdb",
        "Authentication": "SQL",
        "UserName": "dbuserid",
        "Password": "dbpassword",
        "Port": 3306,
        "Encrypt": false
    },
    "Query":"SELECT d.h_entity_l_id AS lid, al.h_name AS lname, d.h_entity_r_id AS rid, ar.h_name AS rname, d.h_dependency AS dep, i.h_impact AS imp FROM h_cmdb_config_items_dependency d LEFT JOIN h_cmdb_assets al ON d.h_entity_l_id = al.h_pk_asset_id LEFT JOIN h_cmdb_assets ar ON d.h_entity_r_id = ar.h_pk_asset_id LEFT JOIN h_cmdb_config_items_impact i ON d.h_entity_l_id = i.h_entity_l_id AND d.h_entity_r_id = i.h_entity_r_id",
    "AssetIdentifier": {
        "Parent": "lname",
        "Child": "rname",
        "Dependency": "dep",
        "Impact":"imp",
        "Hornbill": "Name"
    },
    "DepencencyMapping": {
        "SourceDependency":"HornbillDependency",
        "Runs":"Runs",
        "Runs On":"Runs On",
        "Hosts":"Hosts",
        "Hosted On":"Hosted On",
        "Members":"Members",
        "Member Of":"Member Of"
    },
    "ImpactMapping": {
        "SourceImpact":"HornbillImpact",
        "Low":"Low",
        "Medium":"Medium",
        "High":"High"
    },
    ,
    "RemoveLinks": false,
    "RemoveQuery":"SELECT d.h_entity_l_id AS lid, al.h_name AS lname, d.h_entity_r_id AS rid, ar.h_name AS rname, d.h_dependency AS dep, i.h_impact AS imp FROM h_cmdb_config_items_dependency d LEFT JOIN h_cmdb_assets al ON d.h_entity_l_id = al.h_pk_asset_id LEFT JOIN h_cmdb_assets ar ON d.h_entity_r_id = ar.h_pk_asset_id LEFT JOIN h_cmdb_config_items_impact i ON d.h_entity_l_id = i.h_entity_l_id AND d.h_entity_r_id = i.h_entity_r_id",
    "RemoveAssetIdentifier": {
        "Parent": "lname",
        "Child": "rname",
        "Dependency": "dep",
        "Impact":"imp",
        "Hornbill": "Name",
        "RemoveBothSides": true
    }
}
```

- `APIKey` - a Hornbill API key for a user account with the correct permissions to carry out all of the required API calls
- `InstanceId` - the Hornbill Instance ID (case sensitive)
- `DBConf`
  - `Driver` - the driver to use to connect to the database that holds the asset information:
    - mssql = Microsoft SQL Server (2005 or above)
    - mysql = MySQL Server 4.1+, MariaDB
    - mysql320 = MySQL Server v3.2.0 to v4.0
    - odbc = ODBC Data Source using SQL Server driver
      - When using ODBC as a data source, the `Database`, `UserName`, `Password` and `Query` parameters should be populated accordingly:
        - Database - this should be populated with the  name of the ODBC connection on the PC that is running the tool
        - UserName - this should be the SQL authentication Username to connect to the Database
        - Password - this should be the password for the above username
        - Query - this should be the SQL query to retrieve the asset records
  - `Server` - The address of the SQL server
  - `Database` - The name of the Database to connect to
  - `Authentication` - The tupe of authentication to use to connect to the SQL server. Can be either:
    - Windows - Windows Account authentication, uses the logged-in Windows account to authenticate
    - SQL - uses SQL Server authentication, and requires the Username and Password parameters (below) to be populated
  - `UserName` The username for the SQL database - only used when Authentication is set to SQL: for Windows authentication this field can be left as an empty string
  - `Password` Password for above User Name - only used when Authentication is set to SQL: for Windows authentication this field can be left as an empty string
  - `Port` SQL port
  - `Encrypt` Boolean value to specify wether the connection between the script and the database should be encrypted. NOTE: There is a bug in SQL Server 2008 and below that causes the connection to fail if the connection is encrypted. Only set this to true if your SQL Server has been patched accordingly
- `Query` The basic SQL query to retrieve asset relationship information from the data source
- `AssetIdentifier` - an object containing details to match asset information returned from the `Query`, above, to existing asset records in your Hornbill instance:
  - `Parent` - specifies the column from the above `Query` that holds the Parent asset unique identifier
  - `Child` - specifies the column from the above `Query` that holds the Child asset unique identifier
  - `Dependency` - specifies the column from the above `Query` that holds the value of the Dependency
  - `Impact` - specifies the column from the above `Query` that holds the value of the Impact
  - `Hornbill` - specifies which column to use from the Hornbill asset records to match with the `Parent` and `Child` column output from the `Query`. The following values are supported:
    - `Name` - This will attempt to match the Hornbill asset using the Name field
    - `Tag` - This will attempt to match the Hornbill asset using the Asset Tag field
    - `Description` - This will attempt to match the Hornbill asset using the Description field
- `DependencyMapping` - an object containing properties to match the dependency column output from the `Query` to the available Hornbill dependency values. The property names should be the dependencies as expected from the `Query` output, and their values should be the matching depencency from your Hornbill instance
- `ImpactMapping` - an object containing properties to match the impact column output from the `Query` to the available Hornbill impact values. The property names should be the impacts as expected from the `Query` output, and their values should be the matching impact from your Hornbill instance
- `RemoveLinks` - Boolean true or flalse, defines whether or not to attempt removal of asset reltionship records
- `RemoveQuery` The basic SQL query to retrieve records for asset relationship removal from the data source
- `RemoveAssetIdentifier` - an object containing details to match asset information returned from the `RemovalQuery`, above, to existing asset and relationship records in your Hornbill instance:
  - `Parent` - specifies the column from the above `RemoveQuery` that holds the Parent asset unique identifier
  - `Child` - specifies the column from the above `RemoveQuery` that holds the Child asset unique identifier
  - `Dependency` - specifies the column from the above `RemoveQuery` that holds the value of the Dependency
  - `Impact` - specifies the column from the above `RemoveQuery` that holds the value of the Impact
  - `Hornbill` - specifies which column to use from the Hornbill asset records to match with the `Parent` and `Child` column output from the `RemoveQuery`. The following values are supported:
    - `Name` - This will attempt to match the Hornbill asset using the Name field
    - `Tag` - This will attempt to match the Hornbill asset using the Asset Tag field
    - `Description` - This will attempt to match the Hornbill asset using the Description field
  - `RemoveBothSides` - Boolean true or false, if the links on both sides of the relationship need to be removed

## Execute

### Command Line Parameters

- `file` - Defaults to `conf.json` - Name of the Configuration file to load
- `dryrun` - Defaults to `false` - Set to `true` and the XML for all XMLMC operations will be dumped to the log file, and any CREATE or UPDATE operations will be skipped. This is to aid in debugging the initial connection information.
- `version` - Defaults to `false` - when set to `true`, the tool will output its version number before exiting

## Testing

If you run the application with the argument dryrun=true then no asset relationships will be created or updated, the XML used to create or update will be saved in the log file so you can ensure the data mappings are correct before running the import.

'goDBAssetRelationships.exe -dryrun=true'

## Scheduling

### Windows

You can schedule goDBAssetRelationships.exe to run with any optional command line argument from Windows Task Scheduler:

- Ensure the user account running the task has rights to goDBAssetRelationships.exe and the containing folder.
- Make sure the Start In parameter contains the folder where goDBAssetRelationships.exe resides in otherwise it will not be able to pick up the correct path.

## Logging

All Logging output is saved in the log directory in the same directory as the executable the file name contains the date and time the import was run 'assetRelationships20190925140000.log'
