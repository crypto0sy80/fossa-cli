# `.fossa.yml`

The fossa configuration file can be created manually by running `fossa init` which is recommended.

> Fields prefixed with `*` are optional and determined at runtime from the environment or omitted entirely.
```yaml
version: 1

cli:
* server: https://app.fossa.com
* fetcher: custom
* project: fossa-cli
* api_key: some-key-here
* revision: 1234567
* locator: custom+github.com/fossas/fossa-cli$revision

analyze:
  modules:
    - name: fossa-cli
      type: go
      target: github.com/fossas/fossa-cli/cmd/fossa
      path: cmd/fossa
*     ignore: false
*     options:
        <option>: <value>
```

It is important to mention that the fossa configuration file will often times have multiple modules such as below:

```yaml
version: 2
cli:
  server: https://app.fossa.com
  fetcher: custom
  project: gradle-test
analyze:
  modules:
  - name: data
    type: gradle
    target: 'data:'
    path: .
    option:
      all-configurations: true
  - name: api
    type: gradle
    target: 'api:'
    path: .
```

In the example above, only one of the modules has an option listed. If `fossa analyze` is run, only the `data` module will be run with `all-configurations: true` as expected. However, configuration file's option can be over ruled by command line options. If `fossa analyze --option all-configurations:false` is run, both modules will be run with `all-configurations: false`, including `data`.


## Fields
### `version:`

Specifies the current fossa configuration file version being used. Version 1 is in production and version 2 is still in development.

### `cli:`
#### `server:` (Optional)
Sets the endpoint that the cli will send requests to. This field should only be modified when running a local on-premise instance of `fossa.com`.

Default: `https://app.fossa.com`

#### `fetcher:` (Optional)
Describes the type of project fossa is uploading, there are two options:
- `custom` - Main fetcher option utilized today, specifies a cli build to the server endpoint. 
- `git` (deprecated) - Corresponds to an existing project on Fossa that was importing from a code hosting site. No longer necessary as Fossa.com differentiates between types of builds.

Default: `custom`

#### `project:` (Optional)
Name of the project being analyzed. This is used to construct part of the project locator used to uniquely identify a project uploaded to fossa.com.

Default: Obtained from version control software (VCS) in the directory.

#### `api_key:` (Optional)
Holds a unique Fossa API Key which is used to determine which organization to associate the upload with. Fossa **strongly advises** against including this field to ensure that personal API keys are not committed to your repository. Fossa recommends running `export FOSSA_API_KEY=<your-api-key>` as an alternative. 

Default: Environment variable `FOSSA_API_KEY` will be used.

#### `revision:` (Optional)
Specifies a projects revision. This is used to construct part of the project locator used to uniquely identify a project uploaded to fossa.com. Revision can be thought of as a version number appended to each upload, used to distinguish the analysis of your present day project from all previous analysis.

Default: Obtains the commit sha from the version control software (VCS) located in the directory.

#### `locator:` (Optional)
Manually specify the project locator that is used to identify the unique project on fossa.com. Fossa does not recommend manually setting the locator.

Default: locator is created using fetcher, api_key to find organization ID, project, revision, and other information from the local VCS software.

### `analyze:`

#### `modules:`
Array of modules that will be analyzed in the order they are listed.

#### `name:`
Name of the module being analyzed. This field has no implication on analysis and is for user organization and debugging purposes only.

#### `type:`
Type of module being analyzed. Supported types can be found in the *Configuration* section of each [supported environment's page](../README.md#Supported-Environments) page.

#### `target:`
Build target for the specified module. Target will be used differently depending on which `type` is selected and is most heavily utilized in analysis methods that shell out to a command such as `go list <target>` or `./gradlew target`.

#### `path:`
Path to the root of the project folder or path to the location of the lockfile. Path will be used differently depending on which `type` is selected.

#### `ignore:` (Optional)
If set to `true` this module will be skipped.

Default: `false`

#### `option:` (Optional)
Most options are unique to the type of module being analyzed. Refer to the [supported environments' pages](../README.md#Supported-Environments) for documentation on the full list of options available.

Default: No options.
