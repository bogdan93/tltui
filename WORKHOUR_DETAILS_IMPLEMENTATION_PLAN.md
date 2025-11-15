# WorkhourDetailsModel Implementation Plan

## Overview
Implement WorkhourDetailsModel following the same pattern as ProjectsModel with table, viewport, and CRUD modals.

## Data Structure
```go
type WorkhourDetails struct {
    ID        int
    Name      string
    ShortName string
    IsWork    bool
}
```

## Implementation Steps

### Phase 1: Core Model Setup
- [x] Create `src/models/workhour-details-model.go`
  - [x] Define WorkhourDetailsModel struct with table, viewport, modals
  - [x] Add FetchAllWorkhourDetails() function (or use existing data)
  - [x] Implement NewWorkhourDetailsModel() constructor
  - [x] Set up table with columns: ID, Name, ShortName, IsWork
  - [x] Configure viewport and table styles
  - [x] Implement Init() method
  - [x] Implement Update() method with message handlers
  - [x] Implement View() method
  - [x] Add getSelectedWorkhourDetail() helper method

### Phase 2: Edit Modal
- [x] Create `src/models/workhour-details-edit-modal.go`
  - [x] Define WorkhourDetailsEditModal struct
  - [x] Add fields: ID, NameInput, ShortNameInput, IsWorkToggle, FocusedInput, ErrorMessage
  - [x] Define WorkhourDetailsEditedMsg struct
  - [x] Define WorkhourDetailsEditCanceledMsg struct
  - [x] Implement NewWorkhourDetailsEditModal() constructor
  - [x] Implement Update() method with validation
    - [x] Name required validation
    - [x] ShortName required validation
    - [x] Tab/Shift-Tab navigation (3 inputs)
    - [x] Space to toggle IsWork checkbox
    - [x] Enter to save
    - [x] ESC to cancel
  - [x] Implement View() method
  - [x] Implement dispatch functions
  - [x] Implement updateInputFocus() helper

### Phase 3: Create Modal
- [x] Create `src/models/workhour-details-create-modal.go`
  - [x] Define WorkhourDetailsCreateModal struct
  - [x] Add same fields as edit modal (without ID)
  - [x] Define WorkhourDetailsCreatedMsg struct
  - [x] Define WorkhourDetailsCreateCanceledMsg struct
  - [x] Implement NewWorkhourDetailsCreateModal() constructor
  - [x] Implement Update() method with validation
  - [x] Implement View() method
  - [x] Implement dispatch functions
  - [x] Implement updateInputFocus() helper

### Phase 4: Delete Modal
- [x] Create `src/models/workhour-details-delete-modal.go`
  - [x] Define WorkhourDetailsDeleteModal struct
  - [x] Define WorkhourDetailsDeletedMsg struct
  - [x] Define WorkhourDetailsDeleteCanceledMsg struct
  - [x] Implement NewWorkhourDetailsDeleteModal() constructor
  - [x] Implement Update() method
  - [x] Implement View() method with warning styling
  - [x] Implement dispatch functions

### Phase 5: Integration
- [x] Update `src/models/app-model.go`
  - [x] Add WorkhourDetailsModel field
  - [x] Initialize in main setup
  - [x] Route ModeViewWorkhourDetails to WorkhourDetailsModel
  - [x] Update Update() method to forward to WorkhourDetailsModel
  - [x] Update View() method to render WorkhourDetailsModel

### Phase 6: Testing
- [x] Build and test the application
- [x] Application builds successfully without errors

## Notes
- Follow exact pattern from ProjectsModel
- Use consistent styling and error messages
- IsWork field needs special handling (checkbox/toggle instead of text input)
- Consider using 3-way tab navigation (Name → ShortName → IsWork)
